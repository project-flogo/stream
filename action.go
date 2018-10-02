package stream

import (
	"context"
	"errors"
	"fmt"
	"github.com/project-flogo/fscript"
	"strings"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/exprs"
	"github.com/project-flogo/core/data/mappers"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine/channels"
	"github.com/project-flogo/core/logger"
	"github.com/project-flogo/stream/pipeline"
)

func init() {
	action.Register(&StreamAction{}, &ActionFactory{})
}

var manager *pipeline.Manager
var actionMd = metadata.New(&Settings{})

type Settings struct {
	PipelineURI   string `md:"pipelineURI,required"`
	GroupBy       string `md:"groupBy"`
	OutputChannel string `md:"outputChannel"`
}

type ActionFactory struct {
}

func (f *ActionFactory) Init() error {

	if manager != nil {
		return nil
	}

	scriptExprFactory := fscript.NewExprFactory(pipeline.GetDataResolver())
	exprFactory := exprs.NewExprFactory(pipeline.GetDataResolver(), scriptExprFactory)
	mapperFactory := mappers.NewExprMapperFactory(exprFactory)

	manager = pipeline.NewManager()
	resource.RegisterLoader(pipeline.RESTYPE, pipeline.NewResourceLoader(mapperFactory, pipeline.GetDataResolver()))

	return nil
}

func (f *ActionFactory) New(config *action.Config, resManager *resource.Manager) (action.Action, error) {

	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	streamAction := &StreamAction{}

	if settings.PipelineURI == "" {
		return nil, fmt.Errorf("pipeline URI not specified")
	}

	if strings.HasPrefix(settings.PipelineURI, resource.UriScheme) {

		res := resManager.GetResource(settings.PipelineURI)

		if res != nil {
			def, ok := res.Object().(*pipeline.Definition)
			if !ok {
				return nil, errors.New("unable to resolve pipeline: " + settings.PipelineURI)
			}
			streamAction.definition = def
		} else {
			return nil, errors.New("unable to resolve pipeline: " + settings.PipelineURI)
		}
	} else {
		def, err := manager.GetPipeline(settings.PipelineURI)
		if err != nil {
			return nil, err
		} else {
			if def == nil {
				return nil, errors.New("unable to resolve pipeline: " + settings.PipelineURI)
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

	//note: single pipeline instance for the moment
	inst := pipeline.NewInstance(streamAction.definition, "", settings.GroupBy == "", streamAction.outChannel)
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

func (s *StreamAction) Metadata() *metadata.Metadata {
	return actionMd
}

func (s *StreamAction) IOMetadata() *metadata.IOMetadata {
	return s.ioMetadata
}

func (s *StreamAction) Run(context context.Context, inputs map[string]interface{}, handler action.ResultHandler) error {

	discriminator := ""

	if s.groupBy != "" {
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
