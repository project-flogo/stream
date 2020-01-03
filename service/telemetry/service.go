package telemetry

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/service"
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

var logger = log.ChildLogger(log.RootLogger(), "stream-telemetry")

type PipelineTelemetry struct {
	PipelineId string                 `json:"pipelineId"`
	InstanceId string                 `json:"instanceId"`
	Type       string                 `json:"type"`
	StageId    string                 `json:"stageId,omitempty"`
	Data       map[string]interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	_ = service.RegisterFactory(&ServiceFactory{})
}

type ServiceFactory struct {
}

func (s *ServiceFactory) NewService(config *service.Config) (service.Service, error) {
	ts := &Service{}
	err := ts.init(config.Settings)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

type Service struct {
	server    *http.Server
	clients   map[*websocket.Conn]bool
	broadcast chan *PipelineTelemetry
	stopChan  chan struct{}
}

func (s *Service) Name() string {
	return "stream-telemetry"
}

//DEPRECATED
func (s *Service) Enabled() bool {
	return true
}

func getTelemetryPort() int {
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

func (s *Service) init(settings map[string]interface{}) error {

	port := getTelemetryPort()
	var err error
	sPort, set := settings["port"]
	if set {
		port, err = coerce.ToInt(sPort)
		if err != nil {
			return err
		}
	}

	router := httprouter.New()
	addr := ":" + strconv.Itoa(port)

	router.Handle("GET", "/telemetry", s.wsHandler)

	s.server = &http.Server{Addr: addr, Handler: router}

	support.RegisterTelemetryService(s)

	return nil
}

func (s *Service) Start() error {

	s.clients = make(map[*websocket.Conn]bool)
	s.broadcast = make(chan *PipelineTelemetry)
	s.stopChan = make(chan struct{})

	err := listenAndServe(s.server)
	if err != nil {
		return err
	}
	go s.notify()

	return nil
}

func (s *Service) Stop() error {

	if s.server != nil {

		close(s.stopChan) // tell it to stop

		err := s.server.Shutdown(context.TODO())
		if err != nil {
			return err
		}
		s.server = nil
	}

	return nil
}

func listenAndServe(srv *http.Server) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}


	fullAddr := srv.Addr
	if fullAddr[0] == ':' {
		fullAddr = "0.0.0.0" + srv.Addr
	}

	logger.Infof("Listening on http://%s", fullAddr)


	go func() {
		defer ln.Close()
		// returns ErrServerClosed on graceful close
		if err := srv.Serve(ln); err != http.ErrServerClosed {
			logger.Errorf("ListenAndServe(): %s", err)
		}
	}()

	return nil
}

func (s *Service) PipelineStarted(pipelineId, instanceId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgPipelineStarted, PipelineId: pipelineId, InstanceId: instanceId, Data: data}
	s.broadcast <- t
}

func (s *Service) StageStarted(pipelineId, instanceId, stageId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgStageStarted, PipelineId: pipelineId, InstanceId: instanceId, StageId: stageId, Data: data}
	s.broadcast <- t
}

func (s *Service) StageFinished(pipelineId, instanceId, stageId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgStageFinished, PipelineId: pipelineId, InstanceId: instanceId, StageId: stageId, Data: data}
	s.broadcast <- t
}

func (s *Service) PipelineFinished(pipelineId, instanceId string, data map[string]interface{}) {
	t := &PipelineTelemetry{Type: MsgPipelineFinished, PipelineId: pipelineId, InstanceId: instanceId, Data: data}
	s.broadcast <- t
}

func (s *Service) wsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
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
				logger.Error(err)
				continue
			}

			// send to every client that is currently connected
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					logger.Errorf("Websocket error: %s", err)
					_ = client.Close()
					delete(s.clients, client)
				}
			}
		case <-s.stopChan:
			// stop
			return
		}
	}
}
