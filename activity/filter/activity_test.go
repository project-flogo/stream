package filter

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

	settings := &Settings{Type: "non-zero", ProceedOnlyOnEmit: true}
	//mf := mapper.NewFactory(resolve.GetBasicResolver())
	iCtx := test.NewActivityInitContext(settings, nil)

	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput(ivValue, 2)
	done, err := act.Eval(tc)
	assert.Nil(t, err)
	assert.True(t, done)

	filtered := tc.GetOutput(ovFiltered).(bool)
	value := tc.GetOutput(ovValue)

	assert.False(t, filtered)
	assert.Equal(t, 2, value)

	tc.SetInput(ivValue, 0)
	done, err = act.Eval(tc)
	assert.Nil(t, err)
	assert.False(t, done)

	filtered = tc.GetOutput(ovFiltered).(bool)
	value = tc.GetOutput(ovValue)

	assert.True(t, filtered)
	assert.Equal(t, 0, value)
}

func TestPOOEFalse(t *testing.T) {

	settings := &Settings{Type: "non-zero", ProceedOnlyOnEmit: false}
	//mf := mapper.NewFactory(resolve.GetBasicResolver())
	iCtx := test.NewActivityInitContext(settings, nil)

	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput(ivValue, 2)
	done, err := act.Eval(tc)
	assert.Nil(t, err)
	assert.True(t, done)

	filtered := tc.GetOutput(ovFiltered).(bool)
	value := tc.GetOutput(ovValue)

	assert.False(t, filtered)
	assert.Equal(t, 2, value)

	tc.SetInput(ivValue, 0)
	done, err = act.Eval(tc)
	assert.Nil(t, err)
	assert.True(t, done)

	filtered = tc.GetOutput(ovFiltered).(bool)
	value = tc.GetOutput(ovValue)

	assert.True(t, filtered)
	assert.Equal(t, 0, value)
}
