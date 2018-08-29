package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/flogo-oss/core/activity"
	"github.com/flogo-oss/core/data"
)

type StateManager interface {
	GetState(id string) State
}

type State interface {
	GetScope() data.MutableScope

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
	state := &simpleState{scope: &SharedScope{}}

	return &singelStateManager{state: state}
}

type singelStateManager struct {
	state State
}

func (p *singelStateManager) GetState(id string) State {
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
		state = &simpleState{scope: &SharedScope{}}
		p.states[id] = state
	}

	return state
}

type simpleState struct {
	scope      data.MutableScope
	sharedData map[activity.Activity]map[string]interface{}

	//todo optimize: share tickers closer to instance level (there could even be 1 per difinition, just multiple callbacks)

	tickers map[activity.Activity]*TickerHolder
	timers  map[activity.Activity]*TimerHolder
}

func (p *simpleState) GetScope() data.MutableScope {
	return p.scope
}

func (p *simpleState) GetSharedData(act activity.Activity) map[string]interface{} {
	// create map if it doesn't exist
	if p.sharedData == nil {
		p.sharedData = make(map[activity.Activity]map[string]interface{})
	}

	sd, exists := p.sharedData[act]
	if !exists {
		sd = make(map[string]interface{})
		p.sharedData[act] = sd
	}

	return sd
}

func (p *simpleState) NewTicker(act activity.Activity, interval time.Duration) (*TickerHolder, error) {

	if p.tickers == nil {
		p.tickers = make(map[activity.Activity]*TickerHolder)
	} else {
		_, exists := p.tickers[act]
		if exists {
			return nil, fmt.Errorf("multiple tickers not supported, ticker already exists for this activity")
		}
	}

	ticker := time.NewTicker(interval)
	holder := &TickerHolder{mutex: &sync.RWMutex{}, ticker: ticker}
	p.tickers[act] = holder

	return holder, nil
}

func (p *simpleState) GetTicker(act activity.Activity) (*TickerHolder, bool) {

	if p.tickers == nil {
		return nil, false
	}

	ticker, exists := p.tickers[act]

	return ticker, exists
}

func (p *simpleState) RemoveTicker(act activity.Activity) bool {

	if p.tickers == nil {
		return false
	}

	holder, exists := p.tickers[act]
	if exists {
		holder.ticker.Stop()
		delete(p.tickers, act)
		return true
	}

	return false
}

func (p *simpleState) NewTimer(act activity.Activity, interval time.Duration) (*TimerHolder, error) {

	if p.timers == nil {
		p.timers = make(map[activity.Activity]*TimerHolder)
	} else {
		_, exists := p.tickers[act]
		if exists {
			return nil, fmt.Errorf("multiple timers not supported, timer already exists for this activity")
		}
	}

	timer := time.NewTimer(interval)
	holder := &TimerHolder{mutex: &sync.RWMutex{}, timer: timer}
	p.timers[act] = holder

	return holder, nil
}

func (p *simpleState) GetTimer(act activity.Activity) (*TimerHolder, bool) {

	if p.timers == nil {
		return nil, false
	}

	holder, exists := p.timers[act]

	return holder, exists
}

func (p *simpleState) RemoveTimer(act activity.Activity) bool {

	if p.timers == nil {
		return false
	}

	holder, exists := p.timers[act]
	if exists {
		holder.GetTimer().Stop()
		delete(p.timers, act)
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
