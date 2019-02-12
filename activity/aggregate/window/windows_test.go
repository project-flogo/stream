package window

import (
	"testing"

	"github.com/project-flogo/stream/activity/aggregate/window/functions"
	"github.com/stretchr/testify/assert"
)

//note:  using interface{} 4x slower than using specific types, starting with interface{} for expediency
func BenchmarkTumblingWindow_AddSample(b *testing.B) {

	w := NewTumblingWindow(functions.AddSampleSum, functions.AggregateSingleAvg, &Settings{Size: 3})

	s := []int{5, 3, 2}
	for i := 0; i < b.N; i++ {
		w.AddSample(s)
	}
}

func TestTumblingWindow_AddSample(t *testing.T) {

	w := NewTumblingWindow(functions.AddSampleSum, functions.AggregateSingleAvg, &Settings{Size: 3})

	emit, a := w.AddSample(1)
	assert.False(t, emit)
	emit, a = w.AddSample(2)
	assert.False(t, emit)
	emit, a = w.AddSample(3)
	assert.True(t, emit)
	assert.Equal(t, 2, a)

	emit, a = w.AddSample(4)
	assert.False(t, emit)
	emit, a = w.AddSample(5)
	assert.False(t, emit)
	emit, a = w.AddSample(6)
	assert.True(t, emit)
	assert.Equal(t, 5, a)
}

func TestTumblingWindow_AddSampleAccum(t *testing.T) {

	w := NewTumblingWindow(functions.AddSampleAccum, functions.AggregateSingleNoopFunc, &Settings{Size: 3})

	emit, a := w.AddSample(1)
	assert.False(t, emit)
	emit, a = w.AddSample(2)
	assert.False(t, emit)
	emit, a = w.AddSample(3)
	assert.True(t, emit)

	arr := a.([]interface{})
	assert.Equal(t, 3, len(arr))

	emit, a = w.AddSample(4)
	assert.False(t, emit)
	emit, a = w.AddSample(5)
	assert.False(t, emit)
	emit, a = w.AddSample(6)
	assert.True(t, emit)

	arr = a.([]interface{})
	assert.Equal(t, 3, len(arr))
}

func TestTumblingTimeWindowExt_AddSample(t *testing.T) {

	w := NewTumblingTimeWindow(functions.AddSampleSum, functions.AggregateSingleAvg, &Settings{Size: 10, ExternalTimer: true})

	//block AvgBlock = 3
	w.AddSample(1)
	w.AddSample(2)
	w.AddSample(3)
	w.AddSample(4)
	w.AddSample(5)
	e, v := w.NextBlock()
	assert.True(t, e)
	assert.Equal(t, 3, v)

	//block AvgBlock = 5
	w.AddSample(10)
	w.AddSample(15)
	e, v = w.NextBlock()
	assert.True(t, e)
	assert.Equal(t, 5, v)

	//block AvgBlock = 1
	w.AddSample(4)
	w.AddSample(1)
	e, v = w.NextBlock()
	assert.True(t, e)
	assert.Equal(t, 1, v)
}

func TestTumblingTimeWindowExt_AddAccum(t *testing.T) {

	w := NewTumblingTimeWindow(functions.AddSampleAccum, functions.AggregateSingleNoopFunc, &Settings{Size: 10, ExternalTimer: true})

	//block AvgBlock = 3
	w.AddSample(1)
	w.AddSample(2)
	w.AddSample(3)
	w.AddSample(4)
	w.AddSample(5)
	e, v := w.NextBlock()
	assert.True(t, e)

	arr := v.([]interface{})
	assert.Equal(t, 5, len(arr))

	//block AvgBlock = 5
	w.AddSample(10)
	w.AddSample(15)
	e, v = w.NextBlock()
	assert.True(t, e)

	arr = v.([]interface{})
	assert.Equal(t, 2, len(arr))

	//block AvgBlock = 1
	w.AddSample(4)
	w.AddSample(1)
	e, v = w.NextBlock()
	assert.True(t, e)

	arr = v.([]interface{})
	assert.Equal(t, 2, len(arr))
}

func TestSlidingWindow_AddSample(t *testing.T) {

	w := NewSlidingWindow(functions.AggregateBlocksAvg, &Settings{Size: 5, Resolution: 2})

	emit, a := w.AddSample(1)
	assert.False(t, emit)
	emit, a = w.AddSample(2)
	assert.False(t, emit)
	emit, a = w.AddSample(3)
	assert.False(t, emit)
	emit, a = w.AddSample(4)
	assert.False(t, emit)
	emit, a = w.AddSample(5)
	assert.True(t, emit)
	assert.Equal(t, 3, a)
	emit, a = w.AddSample(6)
	assert.False(t, emit)
	emit, a = w.AddSample(7)
	assert.True(t, emit)
	assert.Equal(t, 5, a)
}

func TestSlidingTimeWindowExt_AddSample(t *testing.T) {

	w := NewSlidingTimeWindow(functions.AddSampleSum, functions.AggregateBlocksAvg, &Settings{Size: 30, Resolution: 10, ExternalTimer: true})

	//block AvgBlock = 3
	w.AddSample(1)
	w.AddSample(2)
	w.AddSample(3)
	w.AddSample(4)
	w.AddSample(5)
	e, v := w.NextBlock()
	assert.False(t, e)

	//block AvgBlock = 2
	w.AddSample(5)
	w.AddSample(5)
	e, _ = w.NextBlock()
	assert.False(t, e)

	//block AvgBlock = 1
	w.AddSample(4)
	w.AddSample(1)
	e, v = w.NextBlock()
	assert.True(t, e)
	assert.Equal(t, 2, v)

	w.AddSample(10)
	w.AddSample(20)
	e, v = w.NextBlock()
	assert.True(t, e)
	assert.Equal(t, 3, v)
}

func TestZero(t *testing.T) {
	var fa []float64
	zero(fa)

	fa2 := make([]float64, 4, 4)
	fa2[0] = 5.5

	zero(fa2)

	zero(nil)
}
