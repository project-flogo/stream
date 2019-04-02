package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
)

type StateManager interface {
	GetState(id string) State
}

type State interface {
	GetScope() data.Scope

	//GetSharedTempData gets the activity instance specific shared data
	GetSharedData(act activity.Activity) map[string]interface{}

	NewTicker(act activity.Activity, interval time.Duration) (*TickerHolder, error)

	GetTicker(act activity.Activity) (*TickerHolder, bool)

	RemoveTicker(act activity.Activity) bool

	NewTimer(act activity.Activity, interval time.Duration) (*TimerHolder, error)

	GetTimer(act activity.Activity) (*TimerHolder, bool)

	RemoveTimer(act activity.Activity) bool
}

func NewSimpleStateManager() StateManager {

	//tickers map[activity.Activity]*time.Ticker
	//todo optimize so only created for activities that need it
	state := &simpleState{scope: &SharedScope{}, sharedData: make(map[activity.Activity]map[string]interface{}), mutex: &sync.RWMutex{}}

	return &singleStateManager{state: state}
}

type singleStateManager struct {
	state State
}

func (p *singleStateManager) GetState(id string) State {
	return p.state
}

func NewMultiStateManager() StateManager {
	return &multiStateManager{states: make(map[string]State)}
}

type multiStateManager struct {
	states map[string]State

	//tickers map[activity.Activity]*time.Ticker
	//repeating timer
	// ticker with slice of callbacks
	// add callback

	rwMutex sync.RWMutex
}

func (p *multiStateManager) GetState(id string) State {

	p.rwMutex.RLock()
	//fast path
	if state, exist := p.states[id]; exist {
		p.rwMutex.RUnlock()
		return state
	}
	p.rwMutex.RUnlock()

	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	state, exist := p.states[id]

	if !exist {
		//todo optimize so only created for activities that need it
		state = &simpleState{scope: &SharedScope{}, sharedData: make(map[activity.Activity]map[string]interface{}), mutex: &sync.RWMutex{}}
		p.states[id] = state
	}

	return state
}

type simpleState struct {
	scope      data.Scope
	sharedData map[activity.Activity]map[string]interface{}

	//todo optimize: share tickers closer to instance level (there could even be 1 per definition, just multiple callbacks)

	tickers map[activity.Activity]*TickerHolder
	timers  map[activity.Activity]*TimerHolder

	mutex *sync.RWMutex
}

func (s *simpleState) GetScope() data.Scope {
	return s.scope
}

func (s *simpleState) GetSharedData(act activity.Activity) map[string]interface{} {

	s.mutex.RLock()
	sd, exists := s.sharedData[act]
	s.mutex.RUnlock()

	if !exists {
		s.mutex.Lock()
		sd, exists = s.sharedData[act]
		if !exists {
			sd = make(map[string]interface{})
			s.sharedData[act] = sd
		}
		s.mutex.Unlock()
	}

	return sd
}

func (s *simpleState) NewTicker(act activity.Activity, interval time.Duration) (*TickerHolder, error) {

	if s.tickers == nil {
		s.tickers = make(map[activity.Activity]*TickerHolder)
	} else {
		_, exists := s.tickers[act]
		if exists {
			return nil, fmt.Errorf("multiple tickers not supported, ticker already exists for this activity")
		}
	}

	ticker := time.NewTicker(interval)
	holder := &TickerHolder{mutex: &sync.RWMutex{}, ticker: ticker}
	s.tickers[act] = holder

	return holder, nil
}

func (s *simpleState) GetTicker(act activity.Activity) (*TickerHolder, bool) {

	if s.tickers == nil {
		return nil, false
	}

	ticker, exists := s.tickers[act]

	return ticker, exists
}

func (s *simpleState) RemoveTicker(act activity.Activity) bool {

	if s.tickers == nil {
		return false
	}

	holder, exists := s.tickers[act]
	if exists {
		holder.ticker.Stop()
		delete(s.tickers, act)
		return true
	}

	return false
}

func (s *simpleState) NewTimer(act activity.Activity, interval time.Duration) (*TimerHolder, error) {

	if s.timers == nil {
		s.timers = make(map[activity.Activity]*TimerHolder)
	} else {
		_, exists := s.tickers[act]
		if exists {
			return nil, fmt.Errorf("multiple timers not supported, timer already exists for this activity")
		}
	}

	timer := time.NewTimer(interval)
	holder := &TimerHolder{mutex: &sync.RWMutex{}, timer: timer}
	s.timers[act] = holder

	return holder, nil
}

func (s *simpleState) GetTimer(act activity.Activity) (*TimerHolder, bool) {

	if s.timers == nil {
		return nil, false
	}

	holder, exists := s.timers[act]

	return holder, exists
}

func (s *simpleState) RemoveTimer(act activity.Activity) bool {

	if s.timers == nil {
		return false
	}

	holder, exists := s.timers[act]
	if exists {
		holder.GetTimer().Stop()
		delete(s.timers, act)
		return true
	}

	return false
}

type TickerHolder struct {
	ticker  *time.Ticker
	execCtx *ExecutionContext
	mutex   *sync.RWMutex
}

func (t *TickerHolder) GetTicker() *time.Ticker {
	return t.ticker
}

func (t *TickerHolder) SetLastExecCtx(ctx *ExecutionContext) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.execCtx = ctx
}

func (t *TickerHolder) GetLastExecCtx() *ExecutionContext {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	ctx := t.execCtx
	t.execCtx = nil

	return ctx
}

type TimerHolder struct {
	timer   *time.Timer
	execCtx *ExecutionContext
	mutex   *sync.RWMutex
}

func (t *TimerHolder) GetTimer() *time.Timer {
	return t.timer
}

func (t *TimerHolder) SetLastExecCtx(ctx *ExecutionContext) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.execCtx = ctx
}

func (t *TimerHolder) GetLastExecCtx() *ExecutionContext {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.execCtx
}
