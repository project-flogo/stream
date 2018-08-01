package pipeline

import (
	"sync"
	"fmt"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"errors"
)

type ScopeProvider interface {
	GetScope(id string) data.MutableScope
}

func NewSingleScopeProvider() ScopeProvider {

	return &singleScopeProvider{scope:&SharedScope{}}
}

type singleScopeProvider struct {
	scope data.MutableScope
}

func (p *singleScopeProvider) GetScope(id string) data.MutableScope  {
	return p.scope
}

func NewMultiScopeProvider() ScopeProvider {
	return &multiScopeProvider{scopes:make(map[string]data.MutableScope)}
}


type multiScopeProvider struct {
	scopes map[string]data.MutableScope
	rwMutex sync.RWMutex
}

func (p *multiScopeProvider) GetScope(id string) data.MutableScope  {

	p.rwMutex.RLock()
	//fast path
	if scope,exist := p.scopes[id];exist {
		p.rwMutex.RUnlock()
		return scope
	}
	p.rwMutex.RUnlock()

	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	scope,exist := p.scopes[id]

	if !exist {
		scope = &SharedScope{}
		p.scopes[id] = scope
	}

	return scope
}


type SharedScope struct {
	attrs map[string]*data.Attribute
	rwMutex sync.RWMutex
}

/////////////////////////////////////////
//  data.Scope Implementation

// GetAttr implements data.Scope.GetAttr
func (inst *SharedScope) GetAttr(attrName string) (value *data.Attribute, exists bool) {

	inst.rwMutex.RLock()
	defer inst.rwMutex.RUnlock()

	if inst.attrs != nil {
		attr, found := inst.attrs[attrName]

		if found {
			return attr, true
		}
	}

	return nil, false
}

// SetAttrValue implements data.Scope.SetAttrValue
func (inst *SharedScope) SetAttrValue(attrName string, value interface{}) error {

	inst.rwMutex.Lock()
	defer inst.rwMutex.Unlock()

	if inst.attrs == nil {
		//todo is it work allocating this lazily?
		inst.attrs = make(map[string]*data.Attribute)
	}

	logger.Debugf("SetAttr - name: %s, value:%v\n", attrName, value)

	existingAttr, exists := inst.GetAttr(attrName)

	if exists {
		existingAttr.SetValue(value)
		return nil
	}

	return fmt.Errorf("Attr [%s] does not exists", attrName)
}

// AddAttr implements data.MutableScope.SetAttrValue
func (inst *SharedScope) AddAttr(attrName string, attrType data.Type, value interface{}) *data.Attribute {

	inst.rwMutex.Lock()
	defer inst.rwMutex.Unlock()

	if inst.attrs == nil {
		inst.attrs = make(map[string]*data.Attribute)
	}

	logger.Debugf("AddAttr - name: %s, type: %s, value:%v", attrName, attrType, value)

	var attr *data.Attribute

	existingAttr, exists := inst.GetAttr(attrName)

	if exists {
		attr = existingAttr
		attr.SetValue(value)
	} else {
		//todo handle error
		attr, _ = data.NewAttribute(attrName, attrType, value)
		inst.attrs[attrName] = attr
	}

	return attr
}

// SimpleScope is a basic implementation of a scope
type SimpleScope struct {
	parentScope data.Scope
	attrs       map[string]*data.Attribute
}

// NewSimpleScope creates a new SimpleScope
func NewSimpleScope(attrs []*data.Attribute, parentScope data.Scope) data.Scope {

	return newSimpleScope(attrs, parentScope)
}

// NewSimpleScope creates a new SimpleScope
func newSimpleScope(attrs []*data.Attribute, parentScope data.Scope) *SimpleScope {

	scope := &SimpleScope{
		parentScope: parentScope,
		attrs:       make(map[string]*data.Attribute),
	}

	for _, attr := range attrs {
		scope.attrs[attr.Name()] = attr
	}

	return scope
}

// NewSimpleScopeFromMap creates a new SimpleScope
func NewSimpleScopeFromMap(attrs map[string]*data.Attribute, parentScope data.Scope) *SimpleScope {

	scope := &SimpleScope{
		parentScope: parentScope,
		attrs:       attrs,
	}

	return scope
}

// GetAttr implements Scope.GetAttr
func (s *SimpleScope) GetAttr(name string) (attr *data.Attribute, exists bool) {

	attr, found := s.attrs[name]

	if found {
		return attr, true
	}

	if s.parentScope != nil {
		return s.parentScope.GetAttr(name)
	}

	return nil, false
}

// SetAttrValue implements Scope.SetAttrValue
func (s *SimpleScope) SetAttrValue(name string, value interface{}) error {

	attr, found := s.attrs[name]

	if found {
		attr.SetValue(value)
		return nil
	}

	return errors.New("attribute not in scope")
}

// AddAttr implements MutableScope.AddAttr
func (s *SimpleScope) AddAttr(name string, valueType data.Type, value interface{}) *data.Attribute {

	attr, found := s.attrs[name]

	if found {
		attr.SetValue(value)
	} else {
		//todo handle error, add error to AddAttr signature
		attr, _ = data.NewAttribute(name, valueType, value)
		s.attrs[name] = attr
	}

	return attr
}