package pipeline

import (
	"fmt"
	"github.com/project-flogo/core/support"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/log"
)

//var (
//	exists = struct{}{}
//)

type Stage struct {
	act activity.Activity

	settings map[string]interface{}

	outputAttrs map[string]interface{}

	inputMapper  mapper.Mapper
	outputMapper mapper.Mapper
}

type StageConfig struct {
	*activity.Config

	Promotions []string `json:"addToPipeline,omitempty"`
}

type initContextImpl struct {
	settings map[string]interface{}
	mFactory mapper.Factory
}

func (ctx *initContextImpl) Settings() map[string]interface{} {
	return ctx.settings
}

func (ctx *initContextImpl) MapperFactory() mapper.Factory {
	return ctx.mFactory
}

func (ctx *initContextImpl) Logger() log.Logger {
	return log.RootLogger()
}

func NewStage(config *StageConfig, mf mapper.Factory, resolver resolve.CompositeResolver) (*Stage, error) {

	if config.Ref == "" && config.Type != "" {
		log.RootLogger().Warnf("stage configuration 'type' deprecated, use 'ref' in the future")
		config.Ref = "#" + config.Type
	}

	if config.Ref == "" {
		return nil, fmt.Errorf("activity not specified for stage")
	}

	if config.Ref[0] == '#' {
		var ok bool
		activityRef := config.Ref
		config.Ref, ok = support.GetAliasRef("activity", activityRef)
		if !ok {
			return nil, fmt.Errorf("activity '%s' not imported", activityRef)
		}
	}

	act := activity.Get(config.Ref)
	if act == nil {
		return nil, fmt.Errorf("unsupported Activity:" + config.Ref)
	}

	f := activity.GetFactory(config.Ref)

	if f != nil {
		initCtx := &initContextImpl{settings: config.Config.Settings, mFactory: mf}
		pa, err := f(initCtx)
		if err != nil {
			return nil, fmt.Errorf("unable to create stage '%s' : %s", config.Ref, err.Error())
		}
		act = pa
	}

	stage := &Stage{}
	stage.act = act

	settingsMd := act.Metadata().Settings

	if len(config.Settings) > 0 && settingsMd != nil {
		stage.settings = make(map[string]interface{}, len(config.Settings))

		for name, value := range config.Settings {

			attr := settingsMd[name]

			if attr != nil {
				//todo handle error
				stage.settings[name] = resolveSettingValue(resolver, name, value)
			}
		}
	}

	inputAttrs := config.Input

	if len(inputAttrs) > 0 {

		inputMapper, err := mf.NewMapper(inputAttrs)
		if err != nil {
			return nil, err
		}

		stage.inputMapper = inputMapper
	}

	outputAttrs := config.Output

	if len(outputAttrs) > 0 {

		outputMapper, err := mf.NewMapper(outputAttrs)
		if err != nil {
			return nil, err
		}

		stage.outputMapper = outputMapper
	}

	return stage, nil
}

func resolveSettingValue(resolver resolve.CompositeResolver, setting string, value interface{}) interface{} {

	strVal, ok := value.(string)

	if ok && len(strVal) > 0 && strVal[0] == '$' {
		v, err := resolver.Resolve(strVal, nil)

		if err == nil {

			logger.Debugf("Resolved setting [%s: %s] to : %v", setting, value, v)
			return v
		}
	}

	return value
}
