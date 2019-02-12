package aggregate

import (
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	ref := activity.GetRef(&Activity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func TestEval(t *testing.T) {

	settings := &Settings{Function: "avg", WindowType: "tumbling", WindowSize: 2, ProceedOnlyOnEmit: true}
	iCtx := test.NewActivityInitContext(settings, nil)

	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())

	tc.SetInput(ivValue, 2)
	done, err := act.Eval(tc)
	assert.False(t, done)
	assert.Nil(t, err)
	assert.Equal(t, false, tc.GetOutput(ovReport))

	tc.SetInput(ivValue, 4)
	done, err = act.Eval(tc)
	assert.True(t, done)
	assert.Nil(t, err)
	assert.Equal(t, true, tc.GetOutput(ovReport))
	assert.Equal(t, 3, tc.GetOutput(ovResult))
}

func TestPOOEFalse(t *testing.T) {

	settings := &Settings{Function: "avg", WindowType: "tumbling", WindowSize: 2, ProceedOnlyOnEmit: false}
	iCtx := test.NewActivityInitContext(settings, nil)

	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())

	tc.SetInput(ivValue, 2)
	done, err := act.Eval(tc)
	assert.True(t, done)
	assert.Nil(t, err)
	assert.Equal(t, false, tc.GetOutput(ovReport))

	tc.SetInput(ivValue, 4)
	done, err = act.Eval(tc)
	assert.True(t, done)
	assert.Nil(t, err)
	assert.Equal(t, true, tc.GetOutput(ovReport))
	assert.Equal(t, 3, tc.GetOutput(ovResult))
}
