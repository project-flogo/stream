package pipeline

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
)

type DefinitionConfig struct {
	id       string
	Name     string               `json:"name"`
	Metadata *metadata.IOMetadata `json:"metadata"`
	Stages   []*StageConfig       `json:"stages"`
}

func NewDefinition(config *DefinitionConfig, mf mapper.Factory, resolver resolve.CompositeResolver) (*Definition, error) {

	def := &Definition{id: config.id, name: config.Name, metadata: config.Metadata}

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
	id       string
	name     string
	stages   []*Stage
	metadata *metadata.IOMetadata
}

// Metadata returns IO metadata for the pipeline
func (d *Definition) Metadata() *metadata.IOMetadata {
	return d.metadata
}

func (d *Definition) Id() string {
	return d.id
}

func (d *Definition) Name() string {
	return d.name
}

func (d *Definition) Cleanup() error {
	for _, stage := range d.stages {
		if !activity.IsSingleton(stage.act) {
			if needsCleanup, ok := stage.act.(support.NeedsCleanup); ok {
				err := needsCleanup.Cleanup()
				if err != nil {
					log.RootLogger().Warnf("Error cleaning up activity '%s' in stream pipeline '%s' : ", activity.GetRef(stage.act), d.name, err)
				}
			}
		}
	}

	return nil
}
