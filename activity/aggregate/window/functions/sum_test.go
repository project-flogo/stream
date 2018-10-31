package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSampleSum(t *testing.T) {

	var x interface{}
	x = AddSampleSum(x, 3)
	x = AddSampleSum(x, 2)
	x = AddSampleSum(x, 1)

	assert.Equal(t, 6, x)
}

func TestAddSampleSumFloat(t *testing.T) {

	var x interface{}
	x = AddSampleSum(x, 3)
	x = AddSampleSum(x, 2)
	x = AddSampleSum(x, 1)

	assert.Equal(t, 6, x)
}

func TestAggregateBlocksSum(t *testing.T) {

	//values - 5 samples/block
	b := []interface{}{5, 10, 15}
	v := AggregateBlocksSum(b, 0, 1)
	assert.Equal(t, 30, v)

	//values
	b = []interface{}{5, 10, 3}
	v = AggregateBlocksSum(b, 0, 1)
	assert.Equal(t, 18, v)
}
