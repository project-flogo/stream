package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSampleMin(t *testing.T) {

	var x interface{}
	x = AddSampleMin(x, 2)
	x = AddSampleMin(x, 7)
	x = AddSampleMin(x, 3)

	assert.Equal(t, 2, x)
}

func TestAggregateBlocksMin(t *testing.T) {

	b := []interface{}{5, 10, 15}
	v := AggregateBlocksMin(b, 0, 0)
	assert.Equal(t, 5, v)

	b = []interface{}{5, 10, 3}
	v = AggregateBlocksMin(b, 0, 1)
	assert.Equal(t, 3, v)
}
