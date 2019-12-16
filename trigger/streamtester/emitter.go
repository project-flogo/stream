package streamtester

import (
	"context"
	"time"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	Resume = iota + 1
	Pause
	Start
	Stop
	Reload
	Kill
)

type Emitter struct {
	emitDelay      time.Duration
	cmd            chan int
	dataSet        DataSet
	logger         log.Logger
	handler        trigger.Handler
	replayData     bool
	getColumnNames bool
}

func NewEmitter(logger log.Logger, handler trigger.Handler) (*Emitter, error) {

	settings := &HandlerSettings{EmitDelay: 100, ReplayData: true}

	err := metadata.MapToStruct(handler.Settings(), settings, true)
	if err != nil {
		return nil, err
	}

	if settings.EmitDelay < 10 {
		settings.EmitDelay = 10
	}

	 ds, err := NewDataSet(settings.FilePath, settings.ReplayData, settings.DataAsMap, settings.AllDataAtOnce)
	if err != nil {
		return nil, err
	}

	emitter := &Emitter{logger: logger, handler: handler, dataSet:ds}
	emitter.emitDelay = time.Duration(settings.EmitDelay) * time.Millisecond
	emitter.replayData = settings.ReplayData
	emitter.getColumnNames = settings.GetColumnNames
	emitter.cmd = make(chan int)

	return emitter, nil
}

func (e *Emitter) Name() string {
	return e.handler.Name()
}

func (e *Emitter) Start() {
	go func() { e.cmd <- Start }() //does this need to be in a goroutine?
}

func (e *Emitter) Stop() {
	go func() { e.cmd <- Stop }() //does this need to be in a goroutine?
}

func (e *Emitter) Pause() {
	go func() { e.cmd <- Pause }() //does this need to be in a goroutine?
}

func (e *Emitter) Resume() {
	go func() { e.cmd <- Resume }() //does this need to be in a goroutine?
}

func (e *Emitter) Kill() {
	go func() { e.cmd <- Kill }() //does this need to be in a goroutine?
}

func (e *Emitter) Reload() {
	go func() {
		e.cmd <- Stop
		e.cmd <- Reload
	}() //does this need to be in a goroutine?
}

func (e *Emitter) Run() {

	go func() {
		timeChan := time.Tick(e.emitDelay)
		status := 0

		for {
			select {
			case <-timeChan:
				if status == 1 {
					//do work
					output := &Output{}
					if e.getColumnNames {
						output.ColumnNames = e.dataSet.ColumnNames()
					}

					done := false
					output.Data, done = e.dataSet.DataPoint()
					if done {
						continue
					}

					e.logger.Debugf("Trigger Output: %#v", output)

					_, err := e.handler.Handle(context.Background(), output)
					if err != nil {
						e.logger.Errorf("Error while executing handler: %#v", err)
						return
					}
				}
			case cmd := <-e.cmd:
				if cmd == Kill {
					return
				}
				if status == 1 {
					if cmd == Stop || cmd == Pause {
						status = 0
						if cmd == Stop {
							e.dataSet.Reset()
							//reset emitter
						}
					}
				} else {
					if cmd == Reload {
						err := e.dataSet.Reload()
						if err != nil {
							e.logger.Error(err)
						}
						continue
					}
					if cmd == Start || cmd == Resume {
						status = 1
					}
				}
			}
		}
	}()
}
