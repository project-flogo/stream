package pipeline

import (
	"github.com/TIBCOSoftware/flogo-lib/core/data"
)

type DefinitionConfig struct {
	Name     string           `json:"name"`
	Metadata *data.IOMetadata `json:"metadata"`
	Stages   []*StageConfig   `json:"stages"`
}

func NewDefinition(config *DefinitionConfig) (*Definition, error) {

	def := &Definition{name: config.Name, metadata: config.Metadata}

	for _, sconfig := range config.Stages {
		stage, err := NewStage(sconfig)

		if err != nil {
			return nil, err
		}

		def.stages = append(def.stages, stage)
	}

	return def, nil
}

type Definition struct {
	name     string
	stages   []*Stage
	metadata *data.IOMetadata
}

// Metadata returns IO metadata for the flow
func (d *Definition) Metadata() *data.IOMetadata {
	return d.metadata
}


//func NewDefinitionBuilder(name string) *DefinitionBuilder {
//	return &DefinitionBuilder{
//		name:     name,
//		metadata: &data.IOMetadata{},
//	}
//}
//
//type DefinitionBuilder struct {
//	name     string
//	stages   []*Stage
//	metadata *data.IOMetadata
//}
//
//func (b *DefinitionBuilder) AddStage(stage *Stage) *DefinitionBuilder {
//	b.stages = append(b.stages, stage)
//	return b
//}
//
//func (b *DefinitionBuilder) SetInput(input map[string]*data.Attribute) *DefinitionBuilder {
//	b.metadata.Input = input
//	return b
//}
//
//func (b *DefinitionBuilder) SetOutput(input map[string]*data.Attribute) *DefinitionBuilder {
//	b.metadata.Input = input
//	return b
//}
//
//func (b *DefinitionBuilder) Build() *Definition {
//	return &Definition{
//		stages:   b.stages,
//		metadata: b.metadata,
//	}
//}
