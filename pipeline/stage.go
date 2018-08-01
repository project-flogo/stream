package pipeline

import (
	"errors"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var (
	exists = struct{}{}
)

// switch to this mapper style in future?
type MapperAlt interface {
	Apply(inputScope data.Scope) (map[string]*data.Attribute, error)
}

type Stage struct {
	act         activity.Activity
	settings    map[string]*data.Attribute
	inputAttrs  map[string]*data.Attribute
	outputAttrs map[string]*data.Attribute

	inputMapper data.Mapper

	// do we need outputMapper instead?
	promote map[string]struct{} //promote these outputs to the pipeline
}

type StageConfig struct {
	Ref         string                 `json:"ref"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	InputAttrs  map[string]interface{} `json:"input,omitempty"`
	OutputAttrs map[string]interface{} `json:"output,omitempty"`

	InputMappings []*data.MappingDef `json:"inputMappings,omitempty"`
	Promotions    []string           `json:"addToPipeline,omitempty"`
}

//todo switch mappings to simple "A=$.blah" or "A=123" or "A=strings.Concat('a','b')"

func NewStage(config *StageConfig) (*Stage, error) {

	if config.Ref == "" {
		return nil, errors.New("Activity not specified for Stage")
	}

	act := activity.Get(config.Ref)
	if act == nil {
		return nil, errors.New("Unsupported Activity:" + config.Ref)
	}

	stage := &Stage{}
	stage.act = act

	if len(config.Settings) > 0 {
		stage.settings = make(map[string]*data.Attribute, len(config.Settings))

		for name, value := range config.Settings {

			attr := act.Metadata().Settings[name]

			if attr != nil {
				//todo handle error
				stage.settings[name], _ = data.NewAttribute(name, attr.Type(), resolveSettingValue(name, value))
			}
		}
	}

	inputAttrs := config.InputAttrs

	if len(inputAttrs) > 0 {
		stage.inputAttrs = make(map[string]*data.Attribute, len(inputAttrs))

		for name, value := range inputAttrs {

			attr := act.Metadata().Input[name]

			if attr != nil {
				//var err error
				//todo handle error
				stage.inputAttrs[name], _ = data.NewAttribute(name, attr.Type(), value)
			}
		}
	}

	outputAttrs := config.OutputAttrs

	if len(outputAttrs) > 0 {

		stage.outputAttrs = make(map[string]*data.Attribute, len(outputAttrs))

		for name, value := range outputAttrs {

			attr := act.Metadata().Output[name]

			if attr != nil {
				//todo handle error
				stage.outputAttrs[name], _ = data.NewAttribute(name, attr.Type(), value)
			}
		}
	}

	if len(config.Promotions) > 0 {
		stage.promote = make(map[string]struct{})

		for _, value := range config.Promotions {
			stage.promote[value] = exists
		}
	}

	if len(config.InputMappings) > 0 {
		stage.inputMapper = mapper.GetFactory().NewUniqueMapper("", &data.MapperDef{Mappings: config.InputMappings}, GetDataResolver())

	}

	return stage, nil
}

func resolveSettingValue(setting string, value interface{}) interface{} {

	strVal, ok := value.(string)

	if ok && len(strVal) > 0 && strVal[0] == '$' {
		v, err := data.GetBasicResolver().Resolve(strVal, nil)

		if err == nil {

			logger.Debugf("Resolved setting [%s: %s] to : %v", setting, value, v)
			return v
		}
	}

	return value
}
