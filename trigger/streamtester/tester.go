package streamtester

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
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
	logger   log.Logger
	settings *Settings
	server   *http.Server
	emitters []*Emitter
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	t.logger = ctx.Logger()

	for _, handler := range ctx.GetHandlers() {

		emitter, err := NewEmitter(ctx.Logger(), handler)
		if err != nil {
			return err
		}

		t.emitters = append(t.emitters, emitter)
	}

	router := httprouter.New()
	router.POST("/tester/resume", t.resumeEmitter())
	router.POST("/tester/pause", t.pauseEmitter())
	router.POST("/tester/start", t.startEmitter())
	router.POST("/tester/stop", t.stopEmitter())
	router.POST("/tester/reload", t.reloadEmitter())

	router.POST("/tester/resume/:id", t.resumeEmitter())
	router.POST("/tester/pause/:id", t.pauseEmitter())
	router.POST("/tester/start/:id", t.startEmitter())
	router.POST("/tester/stop/:id", t.stopEmitter())
	router.POST("/tester/reload/:id", t.reloadEmitter())


	t.server = &http.Server{Addr: ":" + t.settings.Port, Handler: router}

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {

	for _, emitter := range t.emitters {
		emitter.Run()
	}

	t.logger.Debug("Emitters Ready")
	go func() {
		t.logger.Infof("Listening on http://%s", t.server.Addr)
		_ = t.server.ListenAndServe()
	}()

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {

	for _, emitter := range t.emitters {
		emitter.Kill()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := t.server.Shutdown(ctx)
	return err
}

func (t *Trigger) resumeEmitter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitter := t.getEmitter(id)
			emitter.Resume()
			msg = fmt.Sprintf("Emitter '%s' resumed", id)

		} else {
			for _, emitter := range t.emitters {
				emitter.Resume()
			}
			msg = "All emitters resumed"
		}

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (t *Trigger) pauseEmitter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitter := t.getEmitter(id)
			emitter.Pause()
			msg = fmt.Sprintf("Emitter '%s' paused", id)

		} else {
			for _, emitter := range t.emitters {
				emitter.Pause()
			}
			msg = "All emitters paused"
		}

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (t *Trigger) startEmitter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitter := t.getEmitter(id)
			emitter.Start()
			msg = fmt.Sprintf("Emitter '%s' started", id)

		} else {
			for _, emitter := range t.emitters {
				emitter.Start()
			}
			msg = "All emitters started"
		}

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (t *Trigger) stopEmitter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitter := t.getEmitter(id)
			emitter.Stop()
			msg = fmt.Sprintf("Emitter '%s' stopped", id)

		} else {
			for _, emitter := range t.emitters {
				emitter.Stop()
			}
			msg = "All emitters stopped"
		}

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (t *Trigger) reloadEmitter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitter := t.getEmitter(id)
			emitter.Reload()
			msg = fmt.Sprintf("Emitter '%s' reloaded", id)

		} else {
			for _, emitter := range t.emitters {
				emitter.Reload()
			}
			msg = "All emitters reloaded"
		}

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
		return
	}
}


func (t *Trigger) getEmitter(id string) *Emitter {

	for _, emitter := range t.emitters {
		if id == emitter.Name() {
			return emitter
		}
	}
	return nil
}
