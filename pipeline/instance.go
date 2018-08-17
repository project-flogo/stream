package pipeline

import (
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"strings"
	"errors"
)

type Instance struct {
	def    *Definition
	id     string
	status Status

	sm StateManager
	outChannel  chan interface{}
}

func NewInstance(definition *Definition, id string, single bool) *Instance {

	var sm StateManager

	if single {
		sm = NewSimpleStateManager()
	} else {
		sm = NewMultiStateManager()
	}

	return &Instance{def: definition, id: id, sm: sm}
}

func (inst *Instance) Id() string {
	return inst.id
}

//consider a start/stop instance?

func (inst *Instance) Run(discriminator string, input map[string]*data.Attribute) (output map[string]*data.Attribute, status ExecutionStatus, err error) {

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
		return ctx.pipeineOutput, ctx.status,nil
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

	//prevent re-execution of stage?
	stage := ctx.currentStage()

	logger.Debugf("Pipeline[%s] - Executing stage %d", ctx.pipeline.id, ctx.stageId)

	if stage.inputs != nil {

		in := &StageInputScope{execCtx: ctx}
		ctx.currentInput, err = stage.inputs.GetAttrs(in)
		if err != nil {
			return false, err
		}

	}

	//clear previous output
	ctx.currentOutput = make(map[string]*data.Attribute)

	logger.Debugf("Pipeline[%s] - Evaluating Activity", stage.act.Metadata().ID)

	//eval activity/stage
	done, err = stage.act.Eval(ctx)

	if done {
		currentAttrs := ctx.currentOutput

		if stage.outputs != nil {
			in := &StageOutputScope{execCtx: ctx}
			additionalAttrs, err := stage.outputs.GetDetailedAttrs(in)
			if err != nil {
				return false, err
			}
			for name, attr := range additionalAttrs {
				if attr.isNew {
					if strings.HasPrefix(name, "pipeline.") {
						attrName := name[9:]
						if ctx.pipeineOutput == nil {
							ctx.pipeineOutput = make(map[string]*data.Attribute)
						}
						mdAttr := ctx.pipeline.def.metadata.Output[attrName]
						if mdAttr != nil {
							ctx.pipeineOutput[attrName], err = data.NewAttribute(attrName, mdAttr.Type(), attr.Value())
							if err != nil {
								return false, err
							}
						} else {
							return false, errors.New("unknown pipeline output: " + attrName)
						}
						//get the pipeline metadata
					}
				} else {
					currentAttrs[name] = attr.Attribute
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
		ctx.pipeline.outChannel <- ctx.pipeineOutput
	}

	return nil
}

func ResumeCurrentStage(ctx *ExecutionContext) (done bool, err error) {

	stage := ctx.currentStage()

	logger.Debugf("Pipeline[%s] - Resuming stage %d", ctx.pipeline.id, ctx.stageId)

	aact, ok := stage.act.(activity.AsyncActivity)

	//post-eval activity/stage
	if ok {
		done, err = aact.PostEval(ctx, nil)
	}

	return done, err
}
