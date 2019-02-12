package functions

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestAggregateBlocksAccumulate(t *testing.T) {

	//values - 5 samples/block
	b := []interface{}{5, 10, 15}
	v := AggregateBlocksAccumulate(b, 0, 0)

	expected := []interface{}([]interface{}{5, 10, 15})

	assert.Equal(t, expected, v)

	//values
	b = []interface{}{5, 10, 15}
	v = AggregateBlocksAccumulate(b, 1, 0)

	expected = []interface{}([]interface{}{10, 15, 5})
	assert.Equal(t, expected, v)
}

func TestAddSampleAccum(t *testing.T) {

	//values - 5 samples/block
	b := []interface{}{5, 10, 15}

	accum := AddSampleAccum(nil, b)

	c := []interface{}{5, 10, 15}

	accum = AddSampleAccum(accum, c)

	l := accum.([]interface{})

	assert.Equal(t, 2, len(l))

	fmt.Printf("val: %v", l)
}
