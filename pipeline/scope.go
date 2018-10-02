package pipeline

import (
	"errors"
	"sync"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/logger"
)

type MultiScope interface {
	GetValueByScope(scope string, name string) (value interface{}, exists bool)
}

type SharedScope struct {
	attrs   map[string]interface{}
	rwMutex sync.RWMutex
}

func (s *SharedScope) GetValue(name string) (value interface{}, exists bool) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	if s.attrs != nil {
		attr, found := s.attrs[name]

		if found {
			return attr, true
		}
	}

	return nil, false
}

func (s *SharedScope) SetValue(name string, value interface{}) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	if s.attrs == nil {
		//todo is it work allocating this lazily?
		s.attrs = make(map[string]interface{})
	}

	if logger.DebugEnabled() {
		logger.Debugf("SetAttr - name: %s, value:%v\n", name, value)
	}

	s.attrs[name] = value

	return nil
}

// SimpleScope is a basic implementation of a scope
type SimpleScope struct {
	parentScope data.Scope
	attrs       map[string]interface{}
}

// NewSimpleScope creates a new SimpleScope
func NewSimpleScope(attrs map[string]interface{}, parentScope data.Scope) data.Scope {

	return &SimpleScope{attrs: attrs, parentScope: parentScope}
}

func (s *SimpleScope) GetValue(name string) (value interface{}, exists bool) {
	attr, found := s.attrs[name]

	if found {
		return attr, true
	}

	if s.parentScope != nil {
		return s.parentScope.GetValue(name)
	}

	return nil, false
}

func (s *SimpleScope) SetValue(name string, value interface{}) error {
	s.attrs[name] = value
	return nil
}

// SimpleScope is a basic implementation of a scope
type StageInputScope struct {
	execCtx *ExecutionContext
}

func (s *StageInputScope) GetValue(name string) (value interface{}, exists bool) {

	attrs := s.execCtx.currentOutput

	attr, found := attrs[name]

	if found {
		return attr, true
	}

	return attr, found
}

func (s *StageInputScope) SetValue(name string, value interface{}) error {
	return errors.New("read-only scope")
}

func (s *StageInputScope) GetValueByScope(scope string, name string) (value interface{}, exists bool) {

	//on input
	//   get pipeline inputs : $pipeline[in]
	//   get previous stage output : $.

	attrs := s.execCtx.currentOutput

	if scope == "pipeline" {
		attrs = s.execCtx.pipelineInput
	}

	attr, found := attrs[name]

	if found {
		return attr, true
	}

	return attr, found
}

// SimpleScope is a basic implementation of a scope
type StageOutputScope struct {
	execCtx *ExecutionContext
}

func (s *StageOutputScope) GetValue(name string) (value interface{}, exists bool) {
	attrs := s.execCtx.currentOutput

	attr, found := attrs[name]

	if found {
		return attr, true
	}

	return attr, found
}

func (s *StageOutputScope) SetValue(name string, value interface{}) error {
	return errors.New("read-only scope")
}

func (s *StageOutputScope) GetValueByScope(scope string, name string) (value interface{}, exists bool) {
	attrs := s.execCtx.currentOutput

	switch scope {
	case "pipeline":
		attrs = s.execCtx.pipelineInput
	case "input":
		attrs = s.execCtx.currentInput
	}

	attr, found := attrs[name]

	if found {
		return attr, true
	}

	return attr, found
}
