package csvtimer

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func resumeHandler(resume <-chan int) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		go func() { _ = <-resume }()
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte("The Trigger is resumed.."))

		return
	}
}

func pauseHandler(resume chan<- int) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		go func() { resume <- Pause }()
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte("The Trigger is paused.."))

		return
	}
}

func restartHandler(resume chan<- int) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		resume <- Restart

		w.WriteHeader(200)
		w.Write([]byte("The Trigger is Restarted.."))

		return
	}
}
