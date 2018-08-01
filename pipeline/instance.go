package pipeline

import (
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

type Instance struct {
	def    *Definition
	id     string
	status Status

	sc ScopeProvider
}

func NewInstance(definition *Definition, id string, single bool) *Instance {

	var sc ScopeProvider

	if single {
		sc = NewSingleScopeProvider()
	} else {
		sc = NewMultiScopeProvider()
	}

	return &Instance{def:definition, id:id, sc:sc}
}

func (inst *Instance) Id() string {
	return inst.id
}

//consider a start/stop instance?

func (inst *Instance) Run(discriminator string, input map[string]*data.Attribute) (output map[string]*data.Attribute, err error) {

	hasWork := true

	// todo add initial input to execution context
	ctx := &ExecutionContext{discriminator:discriminator, pipeline:inst }

	//todo make this look nicer
	ctx.output = input

	for hasWork {

		hasWork, err = inst.DoStep(ctx)
		if err != nil {
			break
		}
	}

	if ctx.status == ExecStatusCompleted {
		return ctx.input, nil
	}

	if ctx.status == ExecStatusFailed {
		return nil, err
	}

	return nil, nil
}

func (inst *Instance) DoStep(ctx *ExecutionContext) (hasWork bool, err error) {

	hasNext := false

	if ctx.stageId < len(inst.def.stages) {

		//get the stage to work on
		done, err := ExecuteCurrentStage(ctx)
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

	//do input mappings
	if stage.inputMapper != nil {

		logger.Debugf("Pipeline[%s] - Applying InputMapper", ctx.pipeline.id)

		in := data.NewSimpleScopeFromMap(ctx.output, ctx.pipelineScope())
		out := data.NewFixedScope(stage.act.Metadata().Output)
		err := stage.inputMapper.Apply(in, out)
		if err != nil {
			return false, err
		}

		ctx.input = out.GetAttrs()
	} else {
		//todo review this, what should we do if no mapping is specified
		ctx.input = ctx.output
	}

	//clear previous output
	ctx.output = make(map[string]*data.Attribute)

	logger.Debugf("Pipeline[%s] - Evaluating Activity", stage.act.Metadata().ID)

	//eval activity/stage
	done, err = stage.act.Eval(ctx)

	//add attrs to pipeline
	if len(stage.promote) > 0 {
		scope := ctx.pipelineScope()

		for key, value := range ctx.input {
			if _,exists := stage.promote[key]; exists{
				scope.AddAttr("_P." + key, value.Type(), value.Name())
			}
		}
	}

	return done, err
}

