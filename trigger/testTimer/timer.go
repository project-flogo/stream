package testTimer

import (
	"context"
	"encoding/csv"
	"errors"
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
	Resume = iota + 1
	Pause
	Start
	Stop
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
	handlers []*Handler
	logger   log.Logger
	settings *Settings
	router   *httprouter.Router
	ch       chan int
}
type Handler struct {
	handler  trigger.Handler
	settings *HandlerSettings
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	t.ch = make(chan int)

	t.logger = ctx.Logger()

	if t.settings.Control {
		router := httprouter.New()

		//Register For All.
		router.POST("/control/resume/", resumeHandler(t.ch, nil))
		router.POST("/control/pause/", pauseHandler(t.ch, nil))
		router.POST("/control/start/", startHandler(t.ch, nil))
		router.POST("/control/stop/", stopHandler(t.ch, nil))
		for _, handler := range ctx.GetHandlers() {

			handlerSettings := &HandlerSettings{}
			err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)
			if err != nil {
				return err
			}
			handlerSettings.Ch = make(chan int)
			t.handlers = append(t.handlers, &Handler{handler: handler, settings: handlerSettings})
			//Register Individual
			router.POST("/control/resume/"+handlerSettings.Id, resumeHandler(t.ch, handlerSettings))
			router.POST("/control/pause/"+handlerSettings.Id, pauseHandler(t.ch, handlerSettings))
			router.POST("/control/start/"+handlerSettings.Id, startHandler(t.ch, handlerSettings))
			router.POST("/control/stop/"+handlerSettings.Id, stopHandler(t.ch, handlerSettings))
		}

		t.router = router
	}

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {

	for _, handler := range t.handlers {

		go t.start(handler.handler, handler.settings)

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

	return nil
}
func (t *Trigger) start(handler trigger.Handler, settings *HandlerSettings) {

	var stat int
	for {
		triggerData := &Output{}

		stat = getStatus(t, settings, stat)
		if settings.Block {
			data, err := ReadCsv(settings.FilePath)
			if err != nil {
				return
			}

			triggerData = prepareOnceData(data, settings.Header)
		} else {
			dataTemp, err := ReadCsvInterval(settings)
			if err != nil {
				return
			}
			triggerData = prepareRepeatingData(dataTemp, settings)

			settings.Count = settings.Count + 1
		}

		_, err := handler.Handle(context.Background(), triggerData)

		if err != nil {
			return
		}

		repeatInterval, _ := strconv.Atoi(settings.RepeatInterval)

		time.Sleep(time.Duration(repeatInterval) * time.Millisecond)

	}

}

func prepareOnceData(data [][]string, header bool) *Output {
	triggerData := &Output{}
	if header {

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
	return triggerData

}
func prepareRepeatingData(data []string, settings *HandlerSettings) *Output {
	triggerData := &Output{}

	if settings.Header {
		header := settings.Lines[0]
		obj := make(map[string]interface{})

		for i := 0; i < len(data); i++ {
			if num, err := strconv.ParseFloat(data[i], 64); err == nil {
				obj[header[i]] = num
			} else {
				obj[header[i]] = data[i]
			}

		}

		triggerData.Data = obj

	} else {
		triggerData.Data = data
	}
	return triggerData

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
