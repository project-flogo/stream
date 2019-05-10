package window

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Size          int
	Resolution    int
	ExternalTimer bool

	TotalCountModifier int

	NameKey  string
	ValueKey string
}

func (s *Settings) SetAdditionalSettings(as map[string]string) error {

	for key, value := range as {

		switch strings.ToLower(key) {
		case "totalcountmodifier":
			s.TotalCountModifier, _ = strconv.Atoi(value)
		case "namekey":
			s.NameKey = value
		case "valuekey":
			s.ValueKey = value
		}
	}

	return nil
}

///////////////////
// Tumbling Window

func NewTumblingWindow(addFunc AddSampleFunc, aggFunc AggregateSingleFunc, settings *Settings) Window {

	w := &TumblingWindow{addFunc: addFunc, aggFunc: aggFunc, settings: settings, mutex: &sync.Mutex{}}

	if settings.NameKey != "" {
		w.dataMap = newDataMap(settings.NameKey, settings.ValueKey, addFunc, aggFunc)
	}

	return w
}

//note:  using interface{} 4x slower than using specific types, starting with interface{} for expediency
type TumblingWindow struct {
	addFunc  AddSampleFunc
	aggFunc  AggregateSingleFunc
	settings *Settings

	data       interface{}
	numSamples int

	dataMap *MapData

	mutex *sync.Mutex
}

// AddSample implements window.Window.AddSample
func (w *TumblingWindow) AddSample(sample interface{}) (bool, interface{}) {

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.dataMap != nil {

		w.dataMap.addSample(sample)
		w.numSamples++

		if w.numSamples == w.settings.Size {
			w.numSamples = 0
			return true, w.dataMap.extractData()
		}

		return false, nil
	}

	if s, ok := sample.(string); ok {
		sample, _ = coerce.ToFloat64(s)
		//warn
	}

	//sample size should match data size
	w.data = w.addFunc(w.data, sample)
	w.numSamples++

	if w.numSamples == w.settings.Size {
		// aggregate and emit
		val := w.aggFunc(w.data, w.settings.Size)

		w.numSamples = 0
		w.data, _ = zero(w.data)

		return true, val
	}

	return false, nil
}

///////////////////////
// Tumbling Time Window

func NewTumblingTimeWindow(addFunc AddSampleFunc, aggFunc AggregateSingleFunc, settings *Settings) TimeWindow {
	w := &TumblingTimeWindow{addFunc: addFunc, aggFunc: aggFunc, settings: settings, mutex: &sync.Mutex{}}

	if settings.NameKey != "" {
		w.dataMap = newDataMap(settings.NameKey, settings.ValueKey, addFunc, aggFunc)
	}

	return w
}

// TumblingTimeWindow - A tumbling window based on time. Relies on external entity moving window along
// by calling NextBlock at the appropriate time.
//note:  using interface{} 4x slower than using specific types, starting with interface{} for expediency
type TumblingTimeWindow struct {
	addFunc  AddSampleFunc
	aggFunc  AggregateSingleFunc
	settings *Settings

	data       interface{}
	maxSamples int
	numSamples int

	nextEmit int
	lastAdd  int

	mutex *sync.Mutex

	dataMap *MapData
}

func (w *TumblingTimeWindow) AddSample(sample interface{}) (bool, interface{}) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.dataMap != nil {
		w.dataMap.addSample(sample)
	} else {
		if s, ok := sample.(string); ok {
			sample, _ = coerce.ToFloat64(s)
			//warn
		}

		w.data = w.addFunc(w.data, sample)
	}

	w.numSamples++

	if w.numSamples > w.maxSamples {
		w.maxSamples = w.numSamples
	}

	if !w.settings.ExternalTimer {
		currentTime := getTimeMillis()

		//todo what do we do if this greatly exceeds the nextEmit time?
		if currentTime >= w.nextEmit {
			w.nextEmit = +w.settings.Size // size == time in millis
			return w.nextBlock()
		}
	}

	return false, nil
}

func (w *TumblingTimeWindow) NextBlock() (bool, interface{}) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.nextBlock()
}

func (w *TumblingTimeWindow) nextBlock() (bool, interface{}) {

	var val interface{}

	if w.dataMap != nil {
		val = w.dataMap.extractData()
	} else {
		// aggregate and emit
		val = w.aggFunc(w.data, w.maxSamples) //num samples or max samples?
		w.data, _ = zero(w.data)
	}

	w.numSamples = 0

	if w.settings.TotalCountModifier > 0 {
		//local, so reset max samples
		//todo in the future use average of last N 'numSamples' to calculate max
		w.maxSamples = 0
	}

	return true, val
}

///////////////////
// Sliding Window

func NewSlidingWindow(aggFunc AggregateBlocksFunc, settings *Settings) Window {

	w := &SlidingWindow{aggFunc: aggFunc, settings: settings}
	w.blocks = make([]interface{}, settings.Size)
	w.mutex = &sync.Mutex{}

	if settings.NameKey != "" {
		//todo return error
		return nil
	}

	return w
}

//note:  using interface{} 4x slower than using specific types, starting with interface{} for expediency
// todo split external vs on-add timer
type SlidingWindow struct {
	aggFunc  AggregateBlocksFunc
	settings *Settings

	blocks       []interface{}
	numSamples   int
	currentBlock int
	canEmit      bool

	mutex *sync.Mutex
}

// AddSample implements window.Window.AddSample
func (w *SlidingWindow) AddSample(sample interface{}) (bool, interface{}) {

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if s, ok := sample.(string); ok {
		sample, _ = coerce.ToFloat64(s)
		//warn
	}

	//sample size should match data size
	w.blocks[w.currentBlock] = sample //no addSampleFunc required, just tracking all values

	if !w.canEmit {
		if w.currentBlock == w.settings.Size-1 {
			w.canEmit = true
		}
	}

	w.numSamples++

	if w.canEmit && w.numSamples >= w.settings.Resolution {

		val := w.aggFunc(w.blocks, w.currentBlock, 1)

		w.numSamples = 0
		w.currentBlock++

		w.currentBlock = w.currentBlock % w.settings.Size

		return true, val
	}

	w.currentBlock++

	return false, nil
}

//////////////////////
// Sliding Time Window

func NewSlidingTimeWindow(addFunc AddSampleFunc, aggFunc AggregateBlocksFunc, settings *Settings) TimeWindow {

	numBlocks := settings.Size / settings.Resolution

	w := &SlidingTimeWindow{addFunc: addFunc, aggFunc: aggFunc, numBlocks: numBlocks, settings: settings}

	w.blocks = make([]interface{}, numBlocks)
	w.mutex = &sync.Mutex{}

	if settings.NameKey != "" {
		w.dataMap = newBlockMapData(settings.NameKey, settings.ValueKey, numBlocks, addFunc, aggFunc)
	}

	return w
}

// SlidingTimeWindow - A sliding window based on time. Relies on external entity moving window along
// by calling NextBlock at the appropriate time.
// note:  using interface{} 4x slower than using specific types, starting with interface{} for expediency
type SlidingTimeWindow struct {
	addFunc  AddSampleFunc
	aggFunc  AggregateBlocksFunc
	settings *Settings

	numBlocks    int
	blocks       []interface{}
	maxSamples   int
	numSamples   int
	currentBlock int
	canEmit      bool

	nextBlockTime int
	lastAdd       int

	mutex *sync.Mutex

	dataMap *BlockMapData
}

// AddSample implements window.Window.AddSample
func (w *SlidingTimeWindow) AddSample(sample interface{}) (bool, interface{}) {

	w.mutex.Lock()
	defer w.mutex.Lock()

	if w.dataMap != nil {
		w.dataMap.addBlockSample(w.currentBlock, sample)
	} else {
		if s, ok := sample.(string); ok {
			sample, _ = coerce.ToFloat64(s)
			//warn
		}
		//sample size should match data size
		w.blocks[w.currentBlock] = w.addFunc(w.blocks[w.currentBlock], sample)
	}

	w.numSamples++

	if w.numSamples > w.maxSamples {
		w.maxSamples = w.numSamples
	}

	if !w.settings.ExternalTimer {
		currentTime := getTimeMillis()

		if currentTime > w.nextBlockTime {
			w.nextBlockTime += w.settings.Resolution
			return w.nextBlock()
		}

		return false, nil
	}

	return false, nil
}

func (w *SlidingTimeWindow) NextBlock() (bool, interface{}) {

	w.mutex.Lock()
	defer w.mutex.Lock()

	return w.nextBlock()
}

func (w *SlidingTimeWindow) nextBlock() (bool, interface{}) {

	if !w.canEmit {
		if w.currentBlock == w.numBlocks-1 {
			w.canEmit = true
		}
	}

	w.numSamples = 0
	w.currentBlock++

	if w.canEmit {

		// aggregate and emit

		var val interface{}

		if w.dataMap != nil {
			val = w.dataMap.extractData(w.currentBlock)
		} else {
			val = w.aggFunc(w.blocks, w.currentBlock, w.maxSamples)
		}

		w.currentBlock = w.currentBlock % w.numBlocks
		w.blocks[w.currentBlock], _ = zero(w.blocks[w.currentBlock])
		return true, val
	}

	return false, nil
}

///////////////////
// utils

func zero(a interface{}) (interface{}, error) {
	switch x := a.(type) {
	case int:
		return 0, nil
	case float64:
		return 0.0, nil
	case []int:
		for idx := range x {
			x[idx] = 0
		}
		return x, nil
	case []float64:
		for idx := range x {
			x[idx] = 0.0
		}
		return x, nil
	}

	return nil, fmt.Errorf("unsupported type")
}

func getTimeMillis() int {
	now := time.Now()
	nano := now.Nanosecond()
	return nano / 1000000
}

func newDataMap(nameKey, valueKey string, addFunc AddSampleFunc, aggFunc AggregateSingleFunc) *MapData {
	return &MapData{nameKey: nameKey, valueKey: valueKey, addFunc: addFunc, aggFunc: aggFunc, dataMap: make(map[string]*DataInfo)}
}

type MapData struct {
	nameKey  string
	valueKey string
	addFunc  AddSampleFunc
	aggFunc  AggregateSingleFunc

	dataMap map[string]*DataInfo
}

type DataInfo struct {
	count int
	value interface{}
}

func (md *MapData) addSample(sample interface{}) {

	mSample, ok := sample.(map[string]interface{})
	if !ok {
		//log and exist?
		return
	}

	key, _ := coerce.ToString(mSample[md.nameKey])
	info := md.dataMap[key]
	if info == nil {
		info = &DataInfo{}
		md.dataMap[key] = info
	}

	val := mSample[md.valueKey]

	if s, ok := val.(string); ok {
		val, _ = coerce.ToFloat64(s)
		//warn
	}

	info.count = info.count + 1
	info.value = md.addFunc(info.value, val)
}

func (md *MapData) extractData() map[string]interface{} {

	retMap := make(map[string]interface{}, len(md.dataMap))
	for key, info := range md.dataMap {
		retMap[key] = md.aggFunc(info.value, info.count)
	}

	md.dataMap = make(map[string]*DataInfo) //zero map

	return retMap
}

func newBlockMapData(nameKey, valueKey string, numBlocks int, addFunc AddSampleFunc, aggFunc AggregateBlocksFunc) *BlockMapData {
	return &BlockMapData{nameKey: nameKey, valueKey: valueKey, numBlocks: numBlocks, addFunc: addFunc, aggFunc: aggFunc, dataMap: make(map[string][]*BlockInfo)}
}

type BlockMapData struct {
	nameKey   string
	valueKey  string
	numBlocks int
	dataMap   map[string][]*BlockInfo

	addFunc AddSampleFunc
	aggFunc AggregateBlocksFunc
}

type BlockInfo struct {
	count int
	value interface{}
}

func (md *BlockMapData) addBlockSample(blockId int, sample interface{}) {

	mSample, ok := sample.(map[string]interface{})
	if !ok {
		//log and exist?
		return
	}

	key, _ := coerce.ToString(mSample[md.nameKey])
	blocks := md.dataMap[key]
	if blocks == nil {
		blocks = make([]*BlockInfo, md.numBlocks)
		md.dataMap[key] = blocks
	}

	block := blocks[blockId]
	if block == nil {
		block = &BlockInfo{}
		blocks[blockId] = block
	}

	val := mSample[md.valueKey]

	if s, ok := val.(string); ok {
		val, _ = coerce.ToFloat64(s)
		//warn
	}

	block.value = md.addFunc(block.value, val)
	block.count = block.count + 1
}

func (md *BlockMapData) extractData(blockId int) map[string]interface{} {

	retMap := make(map[string]interface{}, len(md.dataMap))
	for key, blockInfos := range md.dataMap {

		blocks := make([]interface{}, len(blockInfos))
		for id, block := range blockInfos {
			blocks[id] = block.value
		}

		id := blockId
		count := blockInfos[blockId].count

		if md.addFunc == nil {
			id = 0
			count = 1
		}

		retMap[key] = md.aggFunc(blocks, id, count)
	}

	blockId = blockId % md.numBlocks

	// zero blocks
	for _, blockInfos := range md.dataMap {
		blockInfos[blockId] = nil
	}

	return retMap
}
