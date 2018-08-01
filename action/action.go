package action

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"

	"github.com/flogo-oss/stream/pipeline"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/app/resource"
)

const (
	StreamActionRef = "github.com/flogo-oss/stream/action"
)

//var idGenerator *util.Generator
var manager *pipeline.Manager

type Stream struct {
	pipelineURI string
	ioMetadata  *data.IOMetadata
	definition  *pipeline.Definition

	inst        *pipeline.Instance
	groupBy     string
}

type Data struct {
	PipelineURI string `json:"pipelineURI"`
	GroupBy     string `json:"groupBy"`
}

//todo fix this
var metadata = &action.Metadata{ID: "github.com/flogo-oss/stream/action"}

func init() {
	action.RegisterFactory(StreamActionRef, &Factory{})
}

type Factory struct {
}

func (f *Factory) Init() error {

	if manager != nil {
		return nil
	}

	manager = pipeline.NewManager()
	resource.RegisterManager(pipeline.RESTYPE_PIPELINE, manager)

	return nil
}

func (f *Factory) New(config *action.Config) (action.Action, error) {

	streamAction := &Stream{}

	var actionData Data
	err := json.Unmarshal(config.Data, &actionData)
	if err != nil {
		return nil, fmt.Errorf("faild to load pipeline action data '%s' error '%s'", config.Id, err.Error())
	}

	if len(actionData.PipelineURI) > 0 {
		streamAction.pipelineURI = actionData.PipelineURI
	} else {
		return nil, fmt.Errorf("pipeline URI not specified")
	}

	def, err := manager.GetPipeline(streamAction.pipelineURI)
	if err != nil {
		return nil, err
	} else {
		if def == nil {
			return nil, errors.New("unable to resolve pipeline: " + streamAction.pipelineURI)
		}
	}

	streamAction.definition = def

	if config.Metadata != nil {
		streamAction.ioMetadata = config.Metadata
	} else {
		streamAction.ioMetadata = def.Metadata()
	}

	//note: single pipeline instance for the moment
	inst := pipeline.NewInstance(def, "", actionData.GroupBy == "")
	streamAction.inst = inst

	return streamAction, nil
}

func (s *Stream) Metadata() *action.Metadata {
	return metadata
}

func (s *Stream) IOMetadata() *data.IOMetadata {
	return s.ioMetadata
}

func (s *Stream) Run(context context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {

	discriminator := ""

	if s.groupBy != "" {
		//note: for now groupings are determined by inputs to the action
		attr, ok := inputs[s.groupBy]
		if  ok {
			discriminator, _ = data.CoerceToString(attr.Value())
		}
	}

	return s.inst.Run(discriminator, inputs)
}
