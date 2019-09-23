package streamtester

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func resumeHandler(t *Trigger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")
		
		if id != "" {
			emitInfo := getEmitterInfoById(id, t)

			if emitInfo != nil {
				go func() { _ = <-emitInfo.Ch }()
				msg = "The trigger with id " + emitInfo.Name + " is resumed"
			}
		} else {
			for _, handler := range t.handlers {
				go func() { _ = <-handler.EmitInfo.Ch }()
			}
			msg = "The triggers are resumed"
		}
		
		
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))

		return
	}
}

func pauseHandler(t *Trigger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitInfo := getEmitterInfoById(id, t)

			if emitInfo != nil {
				go func() { emitInfo.Ch <- Pause }()
				msg = "The trigger with id " + emitInfo.Name + " is paused"
			}
		} else {
			for _, handler := range t.handlers {
				go func() { handler.EmitInfo.Ch  <- Pause }()
			}
			msg = "The triggers are paused"
		}
		

		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))

		return
	}
}

func startHandler(t *Trigger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitInfo := getEmitterInfoById(id, t)

			if emitInfo != nil {
				go func() { emitInfo.Ch <- Start }()
				msg = "The trigger with id " + emitInfo.Name + " is started"
			}
		} else {
			for _, handler := range t.handlers {
				go func() { handler.EmitInfo.Ch  <- Start }()
			}
			msg = "The triggers are started"
		}
		
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
		return
	}
}
func stopHandler(t *Trigger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		msg := ""
		id := params.ByName("id")

		if id != "" {
			emitInfo := getEmitterInfoById(id, t)

			if emitInfo != nil {
				go func() { emitInfo.Ch <- Stop }()
				msg = "The trigger with id " + emitInfo.Name + " is stopped"
			}
		} else {
			for _, handler := range t.handlers {
				go func() { handler.EmitInfo.Ch  <- Stop }()
			}
			msg = "The triggers are stopped"
		}
	
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
		return

	}
}

func getEmitterInfoById(id string, t *Trigger) *HandlerEmitterInfo{
	
	for _, handler := range t.handlers {
		
		if id == handler.EmitInfo.Name {
			return handler.EmitInfo
		}
	}
	return nil
}