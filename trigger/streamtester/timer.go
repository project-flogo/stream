package streamtester

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
	handlers []*Handler
	logger   log.Logger
	settings *Settings
	router   *httprouter.Router
	ch       chan int
}

type Handler struct {
	handler  trigger.Handler
	EmitInfo *HandlerEmitterInfo
	settings *HandlerSettings
}

type HandlerEmitterInfo struct {
	Name  string
	Count int
	Lines [][]string
	Ch    chan int
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	t.ch = make(chan int)

	t.logger = ctx.Logger()

	router := httprouter.New()
	
	router.POST("/tester/resume", resumeHandler(t))
	router.POST("/tester/pause", pauseHandler(t))
	router.POST("/tester/start", startHandler(t))
	router.POST("/tester/stop", stopHandler(t))
	//Register For All.
	router.POST("/tester/resume/:id", resumeHandler(t))
	router.POST("/tester/pause/:id", pauseHandler(t))
	router.POST("/tester/start/:id", startHandler(t))
	router.POST("/tester/stop/:id", stopHandler(t))

	for _, handler := range ctx.GetHandlers() {

		handlerSettings := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return err
		}

		emitInfo := &HandlerEmitterInfo{}
		var handlerName = handler.Name()

		if handlerName != "Handler" {
			//Register Individual
			emitInfo.Name = handlerName
		}
		emitInfo.Ch = make(chan int)

		t.handlers = append(t.handlers, &Handler{handler: handler, settings: handlerSettings, EmitInfo: emitInfo})

	}

	t.router = router

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {

	for _, handler := range t.handlers {

		go t.start(handler.handler, handler.settings, handler.EmitInfo)

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
func (t *Trigger) start(handler trigger.Handler, settings *HandlerSettings, emitInfo *HandlerEmitterInfo) {

	var stat int
	for {
		triggerData := &Output{}

		stat = getStatus(t, emitInfo, stat)
		if settings.Block {
			data, err := ReadCsv(settings.FilePath)
			if err != nil {
				return
			}

			triggerData = prepareOnceData(data, settings.Header)
		} else {
			dataTemp, err := ReadCsvInterval(settings.FilePath, emitInfo)
			if err != nil {
				return
			}
			triggerData = prepareRepeatingData(dataTemp, emitInfo, settings.Header)

			emitInfo.Count = emitInfo.Count + 1
			//triggerData.Data = dataTemp
		}
		
		t.logger.Debug("Data passed to Handler..",triggerData )
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
func prepareRepeatingData(data []string, emitInfo *HandlerEmitterInfo, header bool) *Output {
	triggerData := &Output{}

	if header {
		if emitInfo.Count == 0 {
			return triggerData
		} 
		headerData := emitInfo.Lines[0]
		obj := make(map[string]interface{})

		for i := 0; i < len(data); i++ {
			if num, err := strconv.ParseFloat(data[i], 64); err == nil {
				obj[headerData[i]] = num
			} else {
				obj[headerData[i]] = data[i]
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

func ReadCsvInterval(path string, emitInfo *HandlerEmitterInfo) ([]string, error) {

	if emitInfo.Count == 0 {
		data, err := ReadCsv(path)
		if err != nil {
			return nil, err
		}
		emitInfo.Lines = data

		return emitInfo.Lines[0], nil
	}
	if emitInfo.Count == len(emitInfo.Lines) {
		return nil, errors.New("Done")
	}

	return emitInfo.Lines[emitInfo.Count], nil

}
