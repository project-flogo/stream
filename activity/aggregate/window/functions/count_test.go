package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSampleCount(t *testing.T) {

	var x interface{}
	x = AddSampleCount(x, "first")
	x = AddSampleCount(x, "second")
	x = AddSampleCount(x, "third")

	assert.Equal(t, 3, x)
}
