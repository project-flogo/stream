package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data"
)

const (
	RESTYPE = "pipeline"
)

func NewResourceLoader(mapperFactory data.MapperFactory, resolver data.CompositeResolver) resource.Loader {
	return &ResourceLoader{mapperFactory: mapperFactory, resolver: resolver}
}

type ResourceLoader struct {
	mapperFactory data.MapperFactory
	resolver      data.CompositeResolver
}

func (rl *ResourceLoader) LoadResource(config *resource.Config) (*resource.Resource, error) {

	var pipelineCfgBytes []byte

	pipelineCfgBytes = config.Data

	var defConfig *DefinitionConfig
	err := json.Unmarshal(pipelineCfgBytes, &defConfig)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pipeline resource with id '%s', %s", config.ID, err.Error())
	}

	pDef, err := NewDefinition(defConfig, rl.mapperFactory, rl.resolver)
	if err != nil {
		return nil, err
	}

	return resource.New(RESTYPE, pDef), nil
}
