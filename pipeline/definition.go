package pipeline

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
)

type DefinitionConfig struct {
	Name     string               `json:"name"`
	Metadata *metadata.IOMetadata `json:"metadata"`
	Stages   []*StageConfig       `json:"stages"`
}

func NewDefinition(config *DefinitionConfig, mf data.MapperFactory, resolver data.CompositeResolver) (*Definition, error) {

	def := &Definition{name: config.Name, metadata: config.Metadata}

	for _, sconfig := range config.Stages {
		stage, err := NewStage(sconfig, mf, resolver)

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
	metadata *metadata.IOMetadata
}

// Metadata returns IO metadata for the flow
func (d *Definition) Metadata() *metadata.IOMetadata {
	return d.metadata
}
