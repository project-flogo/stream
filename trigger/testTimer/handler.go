package testTimer

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func resumeHandler(ch <-chan int, settings *HandlerSettings) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		msg := ""
		if settings == nil {
			go func() { _ = <-ch }()
			msg = "The triggers are resumed"
		} else {
			go func() { _ = <-settings.Ch }()
			msg = "The trigger with id "+ settings.Id +" is resumed"
		}
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(msg))

		return
	}	
}

func pauseHandler(ch chan<- int, settings *HandlerSettings) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		msg := ""
		if settings == nil {
			go func() { ch <- Pause }()
			msg = "The triggers are pause"
		} else {
			go func() { settings.Ch <- Pause }()
			msg = "The trigger with id "+ settings.Id +" is paused"
		}
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(msg))
		
		return
	}
}

func startHandler(ch chan<- int, settings *HandlerSettings) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		msg := ""
		if settings == nil {
			ch <- Start
			msg = "The triggers are started"
		} else {
			go func() { settings.Ch <- Start }()
			msg = "The trigger with id "+ settings.Id +" is started"
		}
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(msg))
		return
	}
}
func stopHandler(ch chan<- int, settings *HandlerSettings) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		msg := ""
		if settings == nil {
			ch <- Stop
			msg = "The triggers are stoped"
		}else {
			go func() { settings.Ch <- Stop }()
			msg = "The trigger with id "+ settings.Id +" is stopped"
		}
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(msg))
		return
	
	}
}