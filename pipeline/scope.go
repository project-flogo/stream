package pipeline

import (
	"errors"
	"sync"
)

type ScopeId int

const (
	ScopeDefault ScopeId = iota
	ScopePipeline
	ScopePassthru
)

var scopeNames = [...]string{
	"default",
	"pipeline",
	"passthru",
}

func (t ScopeId) String() string {
	return scopeNames[t]
}

type MultiScope interface {
	GetValueByScope(scopeId ScopeId, name string) (value interface{}, exists bool)
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
		//todo is it worth allocating this lazily?
		s.attrs = make(map[string]interface{})
	}

	if logger.DebugEnabled() {
		logger.Debugf("SetAttr - name: %s, value:%v\n", name, value)
	}

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

func (s *StageInputScope) GetValueByScope(scopeId ScopeId, name string) (value interface{}, exists bool) {

	attrs := s.execCtx.currentOutput

	switch scopeId {
	case ScopePipeline:
		attrs = s.execCtx.pipelineInput
	case ScopePassthru:
		attrs = s.execCtx.passThru
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

func (s *StageOutputScope) GetValueByScope(scopeId ScopeId, name string) (value interface{}, exists bool) {
	attrs := s.execCtx.currentOutput

	switch scopeId {
	case ScopePipeline:
		attrs = s.execCtx.pipelineInput
	case ScopePassthru:
		attrs = s.execCtx.passThru
	}

	attr, found := attrs[name]

	if found {
		return attr, true
	}

	return attr, found
}
