package filter

import (
	"fmt"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

const (
	ivValue    = "value"
	ovFiltered = "filtered"
	ovValue    = "value"
)

//we can generate json from this! - we could also create a "validate-able" object from this
type Settings struct {
	Type              string `md:"type,allowed(non-zero)"`
	ProceedOnlyOnEmit bool
}

type Input struct {
	Value interface{} `md:"value"`
}

type Output struct {
	Filtered bool        `md:"filtered"`
	Value    interface{} `md:"value"`
}

func init() {
	_ = activity.Register(&Activity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{ProceedOnlyOnEmit: true}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	act := &Activity{}

	if s.Type == "non-zero" {
		act.filter = &NonZeroFilter{}
	} else {
		return nil, fmt.Errorf("unsupported filter: '%s'", s.Type)
	}

	act.proceedOnlyOnEmit = s.ProceedOnlyOnEmit

	return act, nil
}

// Activity is an Activity that is used to Filter a message to the console
type Activity struct {
	filter            Filter
	proceedOnlyOnEmit bool
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Filters the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	filter := a.filter
	proceedOnlyOnEmit := a.proceedOnlyOnEmit

	in := ctx.GetInput(ivValue)

	filteredOut := filter.FilterOut(in)

	done = !(proceedOnlyOnEmit && filteredOut)

	err = ctx.SetOutput(ovFiltered, filteredOut)
	if err != nil {
		return false, err
	}
	err = ctx.SetOutput(ovValue, in)
	if err != nil {
		return false, err
	}

	return done, nil
}

type Filter interface {
	FilterOut(val interface{}) bool
}
