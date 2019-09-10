package csvtimer

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/skothari-tibco/scheduler"
)

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

const (
	Resume = iota
	Pause
	Restart
)

func init() {
	trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}

	err := metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}

	return &Trigger{settings: s}, nil
}

type Trigger struct {
	timers   []*scheduler.Job
	handlers []trigger.Handler
	logger   log.Logger
	settings *Settings
	router   *httprouter.Router
	ch       chan int
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	t.ch = make(chan int)

	t.handlers = ctx.GetHandlers()

	t.logger = ctx.Logger()

	if t.settings.Control {
		router := httprouter.New()
		router.GET("/control/resume", resumeHandler(t.ch))
		router.GET("/control/pause", pauseHandler(t.ch))
		router.GET("/control/restart", restartHandler(t.ch))

		t.router = router
	}

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {

	handlers := t.handlers

	for _, handler := range handlers {

		handlerSettings := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return err
		}

		if handlerSettings.RepeatInterval == "" {
			t.scheduleOnce(handler, handlerSettings)
		} else {
			t.scheduleRepeating(handler, handlerSettings)
		}
	}
	if t.router != nil {

		go func() {
			t.logger.Info("Starting Control Server...")
			http.ListenAndServe(":"+t.settings.Port, t.router)
		}()
	}

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {

	for _, timer := range t.timers {

		if timer.IsRunning() {
			timer.Quit <- true
		}
	}

	t.timers = nil

	return nil
}

func (t *Trigger) scheduleOnce(handler trigger.Handler, settings *HandlerSettings) error {

	seconds := 0

	if settings.StartInterval != "" {
		d, err := time.ParseDuration(settings.StartInterval)
		if err != nil {
			return fmt.Errorf("unable to parse start delay: %s", err.Error())
		}

		seconds = int(d.Seconds())
		t.logger.Debugf("Scheduling action to run once in %d seconds", seconds)
	}

	timerJob := scheduler.Every(seconds).Seconds()

	fn := func() {
		t.logger.Debug("Executing \"Once\" timer trigger")

		data, err := ReadCsv(settings.FilePath)

		triggerData := &Output{}

		triggerData.Error = ""
		if settings.Header {

			obj := make(map[string]interface{})
			for i := 0; i < len(data); i++ {
				for j := 0; j < len(data[0]); j++ {
					if num, err := strconv.ParseFloat(data[i][j], 64); err == nil {
						obj[data[0][j]] = num
					} else {
						obj[data[0][j]] = data[i][j]
					}
				}
			}

			triggerData.Data = obj

		} else {
			triggerData.Data = data
		}
		t.logger.Debug("Passing data to handler", triggerData.Data)
		_, err = handler.Handle(context.Background(), triggerData)
		if err != nil {
			t.logger.Error("Error running handler: ", err.Error())
		}

		if timerJob != nil {
			timerJob.Quit <- true
		}
	}

	if seconds == 0 {
		t.logger.Debug("Start delay not specified, executing action immediately")
		fn()
	} else {
		timerJob, err := timerJob.NotImmediately().Run(fn)
		if err != nil {
			t.logger.Error("Error scheduling execute \"once\" timer: ", err.Error())
		}

		t.timers = append(t.timers, timerJob)
	}
	t.Stop()

	return nil
}

func (t *Trigger) scheduleRepeating(handler trigger.Handler, settings *HandlerSettings) error {
	t.logger.Info("Scheduling a repeating timer")
	var header []string
	startSeconds := 0

	repeatInterval, _ := strconv.Atoi(settings.RepeatInterval)
	settings.Count = 0
	t.logger.Info("reapeat", repeatInterval)
	t.logger.Debugf("Scheduling action to repeat every %d seconds", repeatInterval)

	fn := func() {
		t.logger.Debug("Executing \"Repeating\" timer")

		triggerData := &Output{}
		data, err := ReadCsvInterval(settings)
		settings.Count = settings.Count + 1
		if err != nil {
			t.Stop()

		}

		triggerData.Error = ""
		if settings.Header {
			if len(header) == 0 {
				header = data

			} else {
				obj := make(map[string]interface{})

				for i := 0; i < len(data); i++ {
					if num, err := strconv.ParseFloat(data[i], 64); err == nil {
						obj[header[i]] = num
					} else {
						obj[header[i]] = data[i]
					}

				}

				triggerData.Data = obj
			}

		} else {
			triggerData.Data = data
		}

		t.logger.Debug("Passing data to handler", triggerData.Data)

		if triggerData.Data != nil {

			select {
			case stat := <-t.ch:
				if stat == Restart {
					settings.Count = 1
					break
				}
				t.ch <- Resume
			default:
				break
			}

			_, err = handler.Handle(context.Background(), triggerData)
		}

		if err != nil {
			t.logger.Error("Error running handler: ", err.Error())
		}
	}

	if startSeconds == 0 {

		timerJob, err := scheduler.Every(repeatInterval).MilliSeconds().Run(fn)
		if err != nil {
			t.logger.Error("Error scheduling repeating timer: ", err.Error())
		}

		t.timers = append(t.timers, timerJob)
	}

	return nil
}

func ReadCsv(path string) ([][]string, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()

	if err != nil {
		return nil, err
	}

	return lines, nil
}

func ReadCsvInterval(settings *HandlerSettings) ([]string, error) {

	if settings.Count == 0 {
		data, err := ReadCsv(settings.FilePath)
		if err != nil {
			return nil, err
		}
		settings.Lines = data

		return settings.Lines[0], nil
	}
	if settings.Count == len(settings.Lines) {
		return nil, errors.New("Done")
	}

	return settings.Lines[settings.Count], nil

}
