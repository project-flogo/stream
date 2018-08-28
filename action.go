package stream

import (
	"context"
	"errors"
	"fmt"

	"github.com/TIBCOSoftware/flogo-lib/app/resource"
	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine/channels"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/flogo-oss/stream/pipeline"
)

const (
	actionRef = "github.com/flogo-oss/stream"
)

//var idGenerator *util.Generator
var manager *pipeline.Manager

type StreamAction struct {
	ioMetadata *data.IOMetadata
	definition *pipeline.Definition
	outChannel channels.Channel

	inst    *pipeline.Instance
	groupBy string
}

const (
	sPipelineURI   = "pipelineURI"
	sGroupBy       = "groupBy"
	sOutputChannel = "outputChannel"
)

//we can generate json from this! - we could also create a "validate-able" object from this
type Settings struct {
	PipelineURI   string `md:"required"`
	GroupBy       string
	OutputChannel string
}

//todo fix this
var metadata = &action.Metadata{ID: "github.com/flogo-oss/stream/action", Async: true,
	Settings: map[string]*data.Attribute{"pipeline": data.NewZeroAttribute("pipeline", data.TypeString),
		"groupBy":       data.NewZeroAttribute("groupBy", data.TypeString),
		"outputChannel": data.NewZeroAttribute("outputChannel", data.TypeString)}}

func init() {
	action.RegisterFactory(actionRef, &ActionFactory{})
}

type ActionFactory struct {
	metadata *action.Metadata
}

func (f *ActionFactory) Init() error {

	if manager != nil {
		return nil
	}

	manager = pipeline.NewManager()
	resource.RegisterManager(pipeline.RESTYPE_PIPELINE, manager)

	return nil
}

func (f *ActionFactory) New(config *action.Config) (action.Action, error) {

	streamAction := &StreamAction{}
	settings, err := getSettings(config)
	if err != nil {
		return nil, err
	}

	if settings.PipelineURI == "" {
		return nil, fmt.Errorf("pipeline URI not specified")
	}

	def, err := manager.GetPipeline(settings.PipelineURI)
	if err != nil {
		return nil, err
	} else {
		if def == nil {
			return nil, errors.New("unable to resolve pipeline: " + settings.PipelineURI)
		}
	}

	streamAction.definition = def

	if config.Metadata != nil {
		streamAction.ioMetadata = config.Metadata
	} else {
		streamAction.ioMetadata = def.Metadata()
	}

	if settings.OutputChannel != "" {
		ch := channels.Get(settings.OutputChannel)

		if ch == nil {
			return nil, fmt.Errorf("engine channel `%s` not registered", settings.OutputChannel)
		}

		streamAction.outChannel = ch
	}

	//note: single pipeline instance for the moment
	inst := pipeline.NewInstance(def, "", settings.GroupBy == "", streamAction.outChannel)
	streamAction.inst = inst

	return streamAction, nil
}

func (s *StreamAction) Metadata() *action.Metadata {
	return metadata
}

func (s *StreamAction) IOMetadata() *data.IOMetadata {
	return s.ioMetadata
}

func (s *StreamAction) Run(context context.Context, inputs map[string]*data.Attribute, handler action.ResultHandler) error {

	discriminator := ""

	if s.groupBy != "" {
		//note: for now groupings are determined by inputs to the action
		attr, ok := inputs[s.groupBy]
		if ok {
			discriminator, _ = data.CoerceToString(attr.Value())
		}
	}

	logger.Debugf("Running pipeline")

	go func() {

		defer handler.Done()
		retData, status, err := s.inst.Run(discriminator, inputs)

		if err != nil {
			handler.HandleResult(nil, err)
		} else {
			handler.HandleResult(retData, err)
		}

		if s.outChannel != nil && status == pipeline.ExecStatusCompleted {
			s.outChannel.Publish(retData)
		}
	}()

	return nil
}

func getSettings(config *action.Config) (*Settings, error) {

	settings := &Settings{}

	setting, exists := config.Settings[sPipelineURI]
	if exists {
		//this should be done already for us, action can use its metadata to fix this, defaults and all
		val, err := data.CoerceToString(setting)
		if err == nil {
			settings.PipelineURI = val
		}
	} else {
		//throw error if //sPipelineURI is not defined
	}

	setting, exists = config.Settings[sGroupBy]
	if exists {
		//this should be done already for us, action can use its metadata to fix this, defaults and all
		val, err := data.CoerceToString(setting)
		if err == nil {
			settings.GroupBy = val
		}
	} else {
		//throw error if //sPipelineURI is not defined
	}

	setting, exists = config.Settings[sOutputChannel]
	if exists {
		//this should be done already for us, action can use its metadata to fix this, defaults and all
		val, err := data.CoerceToString(setting)
		if err == nil {
			settings.OutputChannel = val
		}
	} else {
		//throw error if //sPipelineURI is not defined
	}

	// settings validation can be done here once activities are created on configuration instead of
	// setting up during runtime

	return settings, nil
}
