package pipeline

import (
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
)

type DefinitionConfig struct {
	Name     string               `json:"name"`
	Metadata *metadata.IOMetadata `json:"metadata"`
	Stages   []*StageConfig       `json:"stages"`
}

func NewDefinition(config *DefinitionConfig, mf mapper.Factory, resolver resolve.CompositeResolver) (*Definition, error) {

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

// Metadata returns IO metadata for the pipeline
func (d *Definition) Metadata() *metadata.IOMetadata {
	return d.metadata
}

func (d *Definition) Name() string {
	return d.name
}
