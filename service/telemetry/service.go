package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/stream/pipeline/support"
)

const (
	EnvTelemetryPort     = "FLOGO_STREAM_TELEMETRY_PORT"
	DefaultTelemetryPort = 9999

	MsgPipelineStarted  = "pipeline-started"
	MsgPipelineFinished = "pipeline-finished"
	MsgStageStarted     = "stage-started"
	MsgStageFinished    = "stage-finished"
)

func init() {
	support.RegisterTelemetryService(&Service{})
}

//GetRunnerWorkers returns the number of workers to use
func GetTelemetryPort() int {
	port := DefaultTelemetryPort
	portEnv := os.Getenv(EnvTelemetryPort)
	if len(portEnv) > 0 {
		i, err := strconv.Atoi(portEnv)
		if err == nil {
			port = i
		}
	}
	return port
}

type PipelineTelemetry struct {
	PipelineId string                 `json:"pipelineId"`
	Type       string                 `json:"type"`
	StageId    string                 `json:"stageId,omitempty"`
	Data       map[string]interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Service struct {
	logger    log.Logger
	server    *http.Server
	clients   map[*websocket.Conn]bool
	broadcast chan *PipelineTelemetry
	stopchan  chan struct{}
}

func (s *Service) Name() string {
	return "stream-telemetry"
}

func (s *Service) Enabled() bool {
	return true
}

func (s *Service) startTelemetryServer() *http.Server {

	port := GetTelemetryPort()

	router := httprouter.New()
	addr := ":" + strconv.Itoa(port)

	router.Handle("GET", "/telemetry", s.wsHandler)

	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Errorf("ListenAndServe(): %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func (s *Service) Start() error {

	s.clients = make(map[*websocket.Conn]bool)
	s.broadcast = make(chan *PipelineTelemetry)
	s.stopchan = make(chan struct{})
	s.server = s.startTelemetryServer()
	go s.notify()

	return nil
}

func (s *Service) Stop() error {

	if s.server != nil {

		close(s.stopchan) // tell it to stop

		err := s.server.Shutdown(context.TODO())
		if err != nil {
			return err
		}

		s.server = nil
	}

	return nil
}

func (s *Service) PipelineStarted(pipelineId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgPipelineStarted, PipelineId: pipelineId, Data: data}
	s.broadcast <- t
}

func (s *Service) StageStarted(pipelineId, stageId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgStageStarted, PipelineId: pipelineId, Data: data}
	s.broadcast <- t
}

func (s *Service) StageFinished(pipelineId, stageId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgStageFinished, PipelineId: pipelineId, Data: data}
	s.broadcast <- t
}

func (s *Service) PipelineFinished(pipelineId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgPipelineFinished, PipelineId: pipelineId, Data: data}
	s.broadcast <- t
}

func (s *Service) wsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error(err)
		return
	}

	// register client
	s.clients[ws] = true
}

func (s *Service) notify() {

	for {
		select {
		case val := <-s.broadcast:

			msg, err := json.Marshal(val)
			if err != nil {
				s.logger.Error(err)
				continue
			}

			// send to every client that is currently connected
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					s.logger.Errorf("Websocket error: %s", err)
					_ = client.Close()
					delete(s.clients, client)
				}
			}
		case <-s.stopchan:
			// stop
			return
		}
	}
}
