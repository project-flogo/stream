package pipeline

import (
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

type Status int

const (
	// StatusNotStarted indicates that the Pipeline has not started
	StatusNotStarted Status = 0

	// StatusActive indicates that the Pipeline is active
	StatusActive Status = 100

	// StatusDone indicates that the Pipeline is done
	StatusDone Status = 500
)

type ExecutionStatus int

const (
	// ExecStatusNotStarted indicates that the Pipeline execution has not started
	ExecStatusNotStarted ExecutionStatus = 0

	// ExecStatusActive indicates that the Pipeline execution is active
	ExecStatusActive ExecutionStatus = 100

	// ExecStatusStalled indicates that the Pipeline execution has stalled
	ExecStatusStalled ExecutionStatus = 400

	// ExecStatusCompleted indicates that the Pipeline execution has been completed
	ExecStatusCompleted ExecutionStatus = 500

	// ExecStatusCancelled indicates that the Pipeline execution has been cancelled
	ExecStatusCancelled ExecutionStatus = 600

	// ExecStatusFailed indicates that the Pipeline execution has failed
	ExecStatusFailed ExecutionStatus = 700
)

type ExecutionContext struct {
	pipeline      *Instance
	discriminator string

	stageId int
	status  ExecutionStatus

	input  map[string]*data.Attribute
	output map[string]*data.Attribute
}

func (eCtx *ExecutionContext) Status() ExecutionStatus {
	return eCtx.status
}

func (eCtx *ExecutionContext) currentStage() *Stage {
	//possibly keep pointer to state in ctx?
	return eCtx.pipeline.def.stages[eCtx.stageId]
}

func (eCtx *ExecutionContext) pipelineScope() data.MutableScope {
	return eCtx.pipeline.sc.GetScope(eCtx.discriminator)
}

/////////////////////////////////////////
//  activity.Host Implementation

func (eCtx *ExecutionContext) ID() string {
	return eCtx.pipeline.id
}

func (eCtx *ExecutionContext) Name() string {
	return eCtx.pipeline.def.name
}

func (eCtx *ExecutionContext) IOMetadata() *data.IOMetadata {
	return eCtx.pipeline.def.metadata
}

func (eCtx *ExecutionContext) Reply(replyData map[string]*data.Attribute, err error) {
	//ignore - not supported by pipeline
}

func (eCtx *ExecutionContext) Return(returnData map[string]*data.Attribute, err error) {
	//ignore - not supported by pipeline
}

func (eCtx *ExecutionContext) WorkingData() data.Scope {
	return eCtx.pipeline.sc.GetScope(eCtx.discriminator)
}

func (eCtx *ExecutionContext) GetResolver() data.Resolver {
	return data.GetBasicResolver()
}

/////////////////////////////////////////
//  activity.Context Implementation

func (eCtx *ExecutionContext) ActivityHost() activity.Host {
	return eCtx
}

func (eCtx *ExecutionContext) GetSetting(setting string) (value interface{}, exists bool) {
	stage := eCtx.currentStage()
	attr, found := stage.settings[setting]
	if found {
		return attr.Value(), true
	}

	return nil, false
}

func (eCtx *ExecutionContext) GetInitValue(key string) (value interface{}, exists bool) {
	//ignore
	return nil, false
}

func (eCtx *ExecutionContext) GetInput(name string) interface{} {

	attr, found := eCtx.input[name]
	if found {
		return attr.Value()
	} else {
		stage := eCtx.currentStage()
		attr, found := stage.inputAttrs[name]
		if found {
			return attr.Value()
		}
	}

	return nil
}

func (eCtx *ExecutionContext) GetOutput(name string) interface{} {
	attr, found := eCtx.output[name]
	if found {
		return attr.Value()
	} else {
		stage := eCtx.currentStage()
		attr, found := stage.inputAttrs[name]
		if found {
			return attr.Value()
		}
	}

	return nil
}

func (eCtx *ExecutionContext) SetOutput(name string, value interface{}) {

	attr, found := eCtx.output[name]
	if found {
		attr.SetValue(value)
	} else {
		//get type from the stages output or existing metadata
		//todo
		attr, _ = data.NewAttribute(name, data.TypeAny, value)
		eCtx.output[name] = attr
	}
}

// DEPRECATED
func (eCtx *ExecutionContext) TaskName() string {
	//ignore
	return ""
}

// DEPRECATED
func (eCtx *ExecutionContext) FlowDetails() activity.FlowDetails {
	//ignore
	return nil
}
