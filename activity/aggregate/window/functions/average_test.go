package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateSingleAvg(t *testing.T) {
	v := AggregateSingleAvg(10, 5)
	assert.Equal(t, 2, v)
}

func TestAggregateBlocksAvg(t *testing.T) {

	//values - 5 samples/block
	b := []interface{}{5, 10, 15}
	v := AggregateBlocksAvg(b, 0, 5)
	assert.Equal(t, 2, v)

	//values
	b = []interface{}{5, 10, 15}
	v = AggregateBlocksAvg(b, 0, 1)
	assert.Equal(t, 10, v)
}
