package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSampleMax(t *testing.T) {

	var x interface{}
	x = AddSampleMax(x, 2)
	x = AddSampleMax(x, 7)
	x = AddSampleMax(x, 3)

	assert.Equal(t, 7, x)
}

func TestAggregateBlocksMax(t *testing.T) {

	b := []interface{}{5, 10, 15}
	v := AggregateBlocksMax(b, 0, 0)
	assert.Equal(t, 15, v)

	b = []interface{}{5, 10, 3}
	v = AggregateBlocksMax(b, 0, 1)
	assert.Equal(t, 10, v)
}
