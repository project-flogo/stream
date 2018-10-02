package pipeline

import (
	"fmt"
	"github.com/project-flogo/core/data/resolvers"

	"github.com/project-flogo/core/data"
)

var resolver = resolvers.NewCompositeResolver(map[string]data.Resolver{
	".":        &resolvers.ScopeResolver{},
	"env":      &resolvers.EnvResolver{},
	"property": &resolvers.PropertyResolver{},
	"input":    &InputResolver{},
	"pipeline": &PipelineResolver{}})

func GetDataResolver() data.CompositeResolver {
	return resolver
}

var resolverInfo = data.NewResolverInfo(false, false)

type PipelineResolver struct {
}

func (r *PipelineResolver) GetResolverInfo() *data.ResolverInfo {
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

var actResolverInfo = data.NewResolverInfo(false, true)

type InputResolver struct {
}

func (r *InputResolver) GetResolverInfo() *data.ResolverInfo {
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
