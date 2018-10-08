package pipeline

import (
	"errors"
	"fmt"
	"github.com/project-flogo/core/data/coerce"
	"runtime/debug"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/engine/channels"
	"github.com/project-flogo/core/support/logger"
)

type Instance struct {
	def    *Definition
	id     string
	status Status

	sm         StateManager
	outChannel channels.Channel
}

func NewInstance(definition *Definition, id string, single bool, outChannel channels.Channel) *Instance {

	var sm StateManager

	if single {
		sm = NewSimpleStateManager()
	} else {
		sm = NewMultiStateManager()
	}

	return &Instance{def: definition, id: id, sm: sm, outChannel: outChannel}
}

func (inst *Instance) Id() string {
	return inst.id
}

//consider a start/stop instance?

func (inst *Instance) Run(discriminator string, input map[string]interface{}) (output map[string]interface{}, status ExecutionStatus, err error) {

	hasWork := true

	ctx := &ExecutionContext{discriminator: discriminator, pipeline: inst}
	ctx.pipelineInput = input

	//pipeline - current output is the input to the next stage
	ctx.currentOutput = input

	for hasWork {

		hasWork, err = inst.DoStep(ctx, false)
		if err != nil {
			break
		}
	}

	if ctx.status == ExecStatusCompleted {
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

		//get the stage to work on
		done := false
		if resume {
			done, err = ResumeCurrentStage(ctx)
		} else {
			done, err = ExecuteCurrentStage(ctx)
		}

		if err != nil {
			logger.Debugf("Pipeline[%s] - Execution failed - Error: %s", ctx.pipeline.id, err.Error())
			ctx.status = ExecStatusFailed
			return false, err
		}

		if !done {
			logger.Debugf("Pipeline[%s] - Partial Execution Completed", ctx.pipeline.id)

			ctx.UpdateTimers()

			//stage has stalled so we are done working
			ctx.status = ExecStatusStalled
			return false, nil
		}

		ctx.stageId++
		if ctx.stageId < len(inst.def.stages) {
			hasNext = true
		} else {
			logger.Debugf("Pipeline[%s] - Execution Completed", ctx.pipeline.id)
			ctx.status = ExecStatusCompleted
		}
	}

	return hasNext, nil
}

func ExecuteCurrentStage(ctx *ExecutionContext) (done bool, err error) {

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

	//clear previous output
	ctx.currentOutput = make(map[string]interface{})

	if logger.DebugEnabled() {
		ref := activity.GetRef(stage.act)
		logger.Debugf("Pipeline[%s] - Evaluating Activity", ref)
	}

	//eval activity/stage
	done, err = stage.act.Eval(ctx)

	if done {
		currentAttrs := ctx.currentOutput

		if stage.outputMapper != nil {
			in := &StageOutputScope{execCtx: ctx}
			results, err := stage.outputMapper.Apply(in)
			if err != nil {
				return false, err
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
							return false, err
						}
					} else {
						return false, errors.New("unknown pipeline output: " + attrName)
					}
					//			//get the pipeline metadata
				} else {
					currentAttrs[name] = value
				}
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

	return done, err
}
