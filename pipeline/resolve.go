package pipeline

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/data/resolve"
)

var pipelineRes = resolve.NewCompositeResolver(map[string]resolve.Resolver{
	".":        &resolve.ScopeResolver{},
	"env":      &resolve.EnvResolver{},
	"property": &property.Resolver{},
	"pipeline": &MultiScopeResolver{scopeId: ScopePipeline},
	"passthru": &MultiScopeResolver{scopeId: ScopePassthru}})

func GetDataResolver() resolve.CompositeResolver {
	return pipelineRes
}

var resolverInfo = resolve.NewResolverInfo(false, false)

type MultiScopeResolver struct {
	scopeId ScopeId
}

func (r *MultiScopeResolver) GetResolverInfo() *resolve.ResolverInfo {
	return resolverInfo
}

func (r *MultiScopeResolver) Resolve(scope data.Scope, itemName, valueName string) (interface{}, error) {

	var value interface{}
	if ms, ok := scope.(MultiScope); ok {

		var exists bool
		value, exists = ms.GetValueByScope(r.scopeId, valueName)
		if !exists {
			return nil, fmt.Errorf("failed to resolve attr: '%s', not found in %s scope", valueName, r.scopeId)
		}
	}

	return value, nil
}
