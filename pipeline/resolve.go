package pipeline

import (
	"fmt"
	"os"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var resolver = &PipelineResolver{}

func GetDataResolver() data.Resolver {
	return resolver
}

type PipelineResolver struct {
}

func (r *PipelineResolver) Resolve(toResolve string, scope data.Scope) (value interface{}, err error) {

	var details *data.ResolutionDetails

	if strings.HasPrefix(toResolve, "$") {
		details, err = data.GetResolutionDetails(toResolve[1:])
	} else {
		return data.SimpleScopeResolve(toResolve, scope)
	}

	if err != nil {
		return nil, err
	}

	if details == nil {
		return nil, fmt.Errorf("unable to determine resolver for %s", toResolve)
	}

	var exists bool

	switch details.ResolverName {
	case "property":
		// Property resolution
		provider := data.GetPropertyProvider()
		value, exists = provider.GetProperty(details.Property + details.Path) //should we add the path and reset it to ""
		if !exists {
			err := fmt.Errorf("failed to resolve Property: '%s', ensure that property is configured in the application", details.Property)
			logger.Error(err.Error())
			return nil, err
		}
	case "env":
		// Environment resolution
		value, exists = os.LookupEnv(details.Property + details.Path)
		if !exists {
			err := fmt.Errorf("failed to resolve Environment Variable: '%s', ensure that variable is configured", details.Property)
			logger.Error(err.Error())
			return "", err
		}
	case "pipeline":
		attr, exists := scope.GetAttr("_P." + details.Property)
		if !exists {
			return nil, fmt.Errorf("failed to resolve attr: '%s', not found in pipeline", details.Property)
		}
		value = attr.Value()
	case ".":
		//Current scope resolution
		attr, exists := scope.GetAttr(details.Property)
		if !exists {
			return nil, fmt.Errorf("failed to resolve attr: '%s', not found in scope", details.Property)
		}
		value = attr.Value()
	default:
		return nil, fmt.Errorf("unsupported resolver: %s", details.ResolverName)
	}

	if details.Path != "" {
		value, err = data.PathGetValue(value, details.Path)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}

	return value, nil
}
