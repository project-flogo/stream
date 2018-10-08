package pipeline

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/resolve"
)

var pipelineRes = resolve.NewCompositeResolver(map[string]resolve.Resolver{
	".":        &resolve.ScopeResolver{},
	"env":      &resolve.EnvResolver{},
	"property": &resolve.PropertyResolver{},
	"input":    &InputResolver{},
	"pipeline": &PipelineResolver{}})

func GetDataResolver() resolve.CompositeResolver {
	return pipelineRes
}

var resolverInfo = resolve.NewResolverInfo(false, false)

type PipelineResolver struct {
}

func (r *PipelineResolver) GetResolverInfo() *resolve.ResolverInfo {
	return resolverInfo
}

func (r *PipelineResolver) Resolve(scope data.Scope, itemName, valueName string) (interface{}, error) {

	var value interface{}
	if ms, ok := scope.(MultiScope); ok {

		var exists bool
		value, exists = ms.GetValueByScope("pipeline", valueName)
		if !exists {
			return nil, fmt.Errorf("failed to resolve attr: '%s', not found in pipeline", valueName)
		}
	}

	return value, nil
}

var actResolverInfo = resolve.NewResolverInfo(false, true)

type InputResolver struct {
}

func (r *InputResolver) GetResolverInfo() *resolve.ResolverInfo {
	return resolverInfo
}

func (r *InputResolver) Resolve(scope data.Scope, itemName, valueName string) (interface{}, error) {
	var value interface{}

	if ms, ok := scope.(MultiScope); ok {
		var exists bool

		value, exists = ms.GetValueByScope("input", valueName)
		if !exists {
			return nil, fmt.Errorf("failed to resolve attr: '%s', not found in input scope", valueName)
		}
	}
	return value, nil
}
