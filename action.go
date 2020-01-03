package stream

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine/channels"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/stream/pipeline"
)

func init() {
	_ = action.Register(&StreamAction{}, &ActionFactory{})
}

var manager *pipeline.Manager
var actionMd = action.ToMetadata(&Settings{})
var logger log.Logger
var idGenerator *support.Generator

type Settings struct {
	StreamURI     string `md:"streamURI"`
	PipelineURI   string `md:"pipelineURI"`
	GroupBy       string `md:"groupBy"`
	OutputChannel string `md:"outputChannel"`
}

type ActionFactory struct {
	resManager *resource.Manager
}

func (f *ActionFactory) Initialize(ctx action.InitContext) error {

	f.resManager = ctx.ResourceManager()

	logger = log.ChildLogger(log.RootLogger(), "stream")

	if manager != nil {
		return nil
	}

	mapperFactory := mapper.NewFactory(pipeline.GetDataResolver())

	if idGenerator == nil {
		idGenerator, _ = support.NewGenerator()
	}

	manager = pipeline.NewManager()
	err := resource.RegisterLoader(pipeline.ResType, pipeline.NewResourceLoader(mapperFactory, pipeline.GetDataResolver()))
	err = resource.RegisterLoader(pipeline.ResTypeOld, pipeline.NewResourceLoader(mapperFactory, pipeline.GetDataResolver()))
	return err
}

func (f *ActionFactory) New(config *action.Config) (action.Action, error) {

	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	if settings.PipelineURI != "" && settings.StreamURI == "" {
		settings.StreamURI = settings.PipelineURI
	}

	if settings.StreamURI == "" {
		return nil, fmt.Errorf("stream URI not specified")
	}

	streamAction := &StreamAction{}

	if strings.HasPrefix(settings.StreamURI, resource.UriScheme) {

		res := f.resManager.GetResource(settings.StreamURI)

		if res != nil {
			def, ok := res.Object().(*pipeline.Definition)
			if !ok {
				return nil, errors.New("unable to resolve stream: " + settings.StreamURI)
			}
			streamAction.definition = def
		} else {
			return nil, errors.New("unable to resolve stream: " + settings.StreamURI)
		}
	} else {
		def, err := manager.GetPipeline(settings.StreamURI)
		if err != nil {
			return nil, err
		} else {
			if def == nil {
				return nil, errors.New("unable to resolve stream: " + settings.StreamURI)
			}
		}
		streamAction.definition = def
	}

	streamAction.ioMetadata = streamAction.definition.Metadata()

	if settings.OutputChannel != "" {
		ch := channels.Get(settings.OutputChannel)

		if ch == nil {
			return nil, fmt.Errorf("engine channel `%s` not registered", settings.OutputChannel)
		}

		streamAction.outChannel = ch
	}

	instId := idGenerator.NextAsString()
	logger.Debug("Creating Stream Instance: ", instId)

	instLogger := logger

	if log.CtxLoggingEnabled() {
		instLogger = log.ChildLoggerWithFields(logger, log.FieldString("pipelineName", streamAction.definition.Name()), log.FieldString("pipelineId", instId))
	}

	//note: single pipeline instance for the moment
	inst := pipeline.NewInstance(streamAction.definition, instId, settings.GroupBy == "", streamAction.outChannel, instLogger)
	streamAction.inst = inst

	return streamAction, nil
}

type StreamAction struct {
	ioMetadata *metadata.IOMetadata
	definition *pipeline.Definition
	outChannel channels.Channel

	inst    *pipeline.Instance
	groupBy string
}

func (s *StreamAction) Info() *action.Info {
	panic("implement me")
}

func (s *StreamAction) Metadata() *action.Metadata {
	return actionMd
}

func (s *StreamAction) IOMetadata() *metadata.IOMetadata {
	return s.ioMetadata
}

func (s *StreamAction) Run(context context.Context, inputs map[string]interface{}, handler action.ResultHandler) error {

	discriminator := ""

	if s.groupBy != "" {
		//navigate input
		//note: for now groupings are determined by inputs to the action
		value, ok := inputs[s.groupBy]
		if ok {
			discriminator, _ = coerce.ToString(value)
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
