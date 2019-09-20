package pipeline

import (
	"errors"
	"fmt"
	"github.com/project-flogo/stream/pipeline/support"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine/channels"
	"github.com/project-flogo/core/support/log"
)

type Instance struct {
	def    *Definition
	id     string
	status Status
	logger log.Logger

	sm         StateManager
	outChannel channels.Channel
}

func NewInstance(definition *Definition, id string, single bool, outChannel channels.Channel, logger log.Logger) *Instance {

	var sm StateManager

	if single {
		sm = NewSimpleStateManager()
	} else {
		sm = NewMultiStateManager()
	}

	return &Instance{def: definition, id: id, sm: sm, outChannel: outChannel, logger: logger}
}

func (inst *Instance) Id() string {
	return inst.id
}

func (inst *Instance) PipelineId() string {
	return inst.def.Id()
}

//consider a start/stop instance?

func (inst *Instance) Run(discriminator string, input map[string]interface{}) (output map[string]interface{}, status ExecutionStatus, err error) {

	hasWork := true

	//if context logging enable, need to come up with a unique id for the execution
	ctx := &ExecutionContext{discriminator: discriminator, pipeline: inst}
	ctx.pipelineInput = input

	//pipeline - current output is the input to the next stage
	ctx.currentOutput = input

	if t := support.GetTelemetryService(); t != nil {
		t.PipelineStarted(inst.PipelineId(), inst.id, input)
	}

	for hasWork {

		hasWork, err = inst.DoStep(ctx, false)
		if err != nil {
			break
		}
	}

	if ctx.status == ExecStatusCompleted {
		if t := support.GetTelemetryService(); t != nil {
			t.PipelineFinished(inst.PipelineId(), inst.id, ctx.pipelineOutput)
		}

		return ctx.pipelineOutput, ctx.status, nil
	}

	if ctx.status == ExecStatusFailed {
		return nil, ctx.status, err
	}

	return nil, status, nil
}

func (inst *Instance) DoStep(ctx *ExecutionContext, resume bool) (hasWork bool, err error) {

	hasNext := false

	if ctx.stageId < len(inst.def.stages) {

		//if t := support.GetTelemetryService(); t != nil {
		//	t.StageStarted(inst.PipelineId(), inst.id, strconv.Itoa(ctx.stageId), ctx.currentInput)
		//}

		//get the stage to work on
		done := false
		if resume {
			done, err = ResumeCurrentStage(ctx)
		} else {
			done, err = ExecuteCurrentStage(ctx)
		}

		if t := support.GetTelemetryService(); t != nil  && done {
			t.StageFinished(inst.PipelineId(), inst.id, strconv.Itoa(ctx.stageId), ctx.currentOutput)
		}

		if err != nil {
			inst.logger.Debugf("Pipeline[%s] - Execution failed - Error: %s", ctx.pipeline.id, err.Error())
			ctx.status = ExecStatusFailed
			return false, err
		}

		if !done {
			inst.logger.Debugf("Pipeline[%s] - Partial Execution Completed", ctx.pipeline.id)

			ctx.UpdateTimers()

			//stage has stalled so we are done working
			ctx.status = ExecStatusStalled
			return false, nil
		}

		ctx.stageId++
		if ctx.stageId < len(inst.def.stages) {
			hasNext = true
		} else {
			inst.logger.Debugf("Pipeline[%s] - Execution Completed", ctx.pipeline.id)
			ctx.status = ExecStatusCompleted
		}
	}

	return hasNext, nil
}

func ExecuteCurrentStage(ctx *ExecutionContext) (done bool, err error) {

	logger := ctx.pipeline.logger

	defer func() {
		if r := recover(); r != nil {

			err = fmt.Errorf("unhandled error executing stage '%s' : %v", "stage", r)
			logger.Error(err)

			// todo: useful for debugging
			logger.Debugf("StackTrace: %s", debug.Stack())

			done = false
		}
	}()

	//prevent re-execution of stage?
	stage := ctx.currentStage()

	if logger.DebugEnabled() {
		logger.Debugf("Pipeline[%s] - Executing stage %d", ctx.pipeline.id, ctx.stageId)
	}

	if stage.inputMapper != nil {

		in := &StageInputScope{execCtx: ctx}
		ctx.currentInput, err = stage.inputMapper.Apply(in)
		if err != nil {
			return false, err
		}
	}

	if t := support.GetTelemetryService(); t != nil {
		t.StageStarted(ctx.pipeline.PipelineId(), ctx.pipeline.id, strconv.Itoa(ctx.stageId), ctx.currentInput)
	}

	//clear previous output
	ctx.currentOutput = make(map[string]interface{})

	if logger.DebugEnabled() {
		ref := activity.GetRef(stage.act)
		logger.Debugf("Pipeline[%s] - Evaluating Activity", ref)
	}

	//eval activity/stage
	done, err = stage.act.Eval(ctx)

	if done {
		if stage.outputMapper != nil {
			err := applyOutputMapper(ctx)
			if err != nil {
				return false, err
			}
		}
	}

	return done, err
}

func Resume(ctx *ExecutionContext) error {

	hasWork := true

	inst := ctx.pipeline
	var err error

	resume := true
	for hasWork {

		hasWork, err = inst.DoStep(ctx, resume)
		if err != nil {
			break
		}
		resume = false
	}

	if ctx.status == ExecStatusCompleted && ctx.pipeline.outChannel != nil {
		ctx.pipeline.outChannel.Publish(ctx.pipelineOutput)
	}

	return nil
}

func ResumeCurrentStage(ctx *ExecutionContext) (done bool, err error) {

	logger := ctx.pipeline.logger

	defer func() {
		if r := recover(); r != nil {

			err = fmt.Errorf("unhandled Error resuming stage '%s' : %v", "stage", r)
			logger.Error(err)

			// todo: useful for debugging
			logger.Debugf("StackTrace: %s", debug.Stack())

			done = false
		}
	}()

	stage := ctx.currentStage()

	logger.Debugf("Pipeline[%s] - Resuming stage %d", ctx.pipeline.id, ctx.stageId)

	aact, ok := stage.act.(activity.AsyncActivity)

	//post-eval activity/stage
	if ok {
		done, err = aact.PostEval(ctx, nil)
	}

	if done {
		if stage.outputMapper != nil {
			err := applyOutputMapper(ctx)
			if err != nil {
				return false, err
			}
		}
	}

	return done, err
}

func applyOutputMapper(ctx *ExecutionContext) error {

	currentAttrs := ctx.currentOutput
	stage := ctx.currentStage()

	in := &StageOutputScope{execCtx: ctx}
	results, err := stage.outputMapper.Apply(in)
	if err != nil {
		return err
	}

	outputMd := ctx.pipeline.def.metadata.Output
	for name, value := range results {
		if strings.HasPrefix(name, "pipeline.") {
			attrName := name[9:]
			if ctx.pipelineOutput == nil {
				ctx.pipelineOutput = make(map[string]interface{})
			}
			mdAttr := outputMd[attrName]
			if mdAttr != nil {
				ctx.pipelineOutput[attrName], err = coerce.ToType(value, mdAttr.Type())
				if err != nil {
					return err
				}
			} else {
				return errors.New("unknown pipeline output: " + attrName)
			}
			//get the pipeline metadata
		} else if strings.HasPrefix(name, "passthru.") {
			attrName := name[9:]
			if ctx.passThru == nil {
				ctx.passThru = make(map[string]interface{})
			}
			ctx.passThru[attrName] = value
		} else {
			currentAttrs[name] = value
		}
	}

	return nil
}
