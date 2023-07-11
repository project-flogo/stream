package aggregate

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/stream/activity/aggregate/window"
	"github.com/project-flogo/stream/pipeline/support"
)

const (
	ivValue = "value"

	ovResult = "result"
	ovReport = "report"

	sdWindow = "window"
)

//we can generate json from this! - we could also create a "validate-able" object from this
type Settings struct {
	Function           string `md:"function,required,allowed(avg,sum,min,max,count,accumulate)"`
	WindowType         string `md:"windowType,required,allowed(tumbling,sliding,timeTumbling,timeSliding)"`
	WindowSize         int    `md:"windowSize,required"`
	Resolution         int    `md:"resolution"`
	ProceedOnlyOnEmit  bool   `md:"proceedOnlyOnEmit"`
	AdditionalSettings string `md:"additionalSettings"`
}

type Input struct {
	Value interface{} `md:"value"`
}

type Output struct {
	Report bool        `md:"report"`
	Result interface{} `md:"result"`
}

func init() {
	_ = activity.Register(&Activity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{ProceedOnlyOnEmit: true, Resolution: 1}

	//settings.Function = "avg" // default function
	//settings.WindowType = "tumbling" // default window type
	//settings.WindowSize = 5 // default window resolution

	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	additionalSettings, err := toParams(s.AdditionalSettings)
	if err != nil {
		return nil, err
	}

	sharedData := make(map[string]interface{})
	act := &Activity{settings: s, additionalSettings: additionalSettings, sharedData: sharedData}

	return act, nil
}

// Activity is an Activity that is used to Aggregate a message to the console
type Activity struct {
	settings           *Settings
	additionalSettings map[string]string
	mutex              sync.Mutex
	sharedData         map[string]interface{}
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Aggregates the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	sharedData := a.sharedData
	wv, defined := sharedData[sdWindow]

	timerSupport, timerSupported := support.GetTimerSupport(ctx)

	var w window.Window

	//create the window & associated timer if necessary

	if !defined {

		a.mutex.Lock()

		wv, defined = sharedData[sdWindow]
		if defined {
			w = wv.(window.Window)
		} else {
			w, err = a.createWindow(ctx)

			if err != nil {
				a.mutex.Unlock()
				return false, err
			}

			sharedData[sdWindow] = w
		}

		a.mutex.Unlock()
	} else {
		w = wv.(window.Window)
	}

	in := ctx.GetInput(ivValue)

	emit, result := w.AddSample(in)

	if timerSupported {
		timerSupport.UpdateTimer(true)
	}

	err = ctx.SetOutput(ovResult, result)
	if err != nil {
		return false, err
	}

	err = ctx.SetOutput(ovReport, emit)
	if err != nil {
		return false, err
	}

	done = !(a.settings.ProceedOnlyOnEmit && !emit)

	return done, nil
}

func (a *Activity) createWindow(ctx activity.Context) (w window.Window, err error) {

	settings := a.settings
	timerSupport, timerSupported := support.GetTimerSupport(ctx)

	windowSettings := &window.Settings{Size: settings.WindowSize, ExternalTimer: timerSupported, Resolution: settings.Resolution}
	err = windowSettings.SetAdditionalSettings(a.additionalSettings)
	if err != nil {
		return nil, err
	}

	wType := strings.ToLower(settings.WindowType)

	switch wType {
	case "tumbling":
		w, err = NewTumblingWindow(settings.Function, windowSettings)
	case "sliding":
		w, err = NewSlidingWindow(settings.Function, windowSettings)
	case "timetumbling":
		w, err = NewTumblingTimeWindow(settings.Function, windowSettings)
		if err == nil && timerSupported {
			err = timerSupport.CreateTimer(time.Duration(settings.WindowSize)*time.Millisecond, a.moveWindow, true)
		}
	case "timesliding":
		w, err = NewSlidingTimeWindow(settings.Function, windowSettings)
		if err == nil && timerSupported {
			err = timerSupport.CreateTimer(time.Duration(settings.Resolution)*time.Millisecond, a.moveWindow, true)
		}
	default:
		return nil, fmt.Errorf("unsupported window type: '%s'", settings.WindowType)
	}

	return w, err
}

func (a *Activity) PostEval(ctx activity.Context, userData interface{}) (done bool, err error) {
	return true, nil
}

func (a *Activity) moveWindow(ctx activity.Context) bool {

	sharedData := a.sharedData

	wv, _ := sharedData[sdWindow]

	w, _ := wv.(window.TimeWindow)

	emit, result := w.NextBlock()

	err := ctx.SetOutput(ovResult, result)
	if err != nil {
		//todo log error?
	}

	err = ctx.SetOutput(ovReport, emit)
	if err != nil {
		//todo log error?
	}

	return !(a.settings.ProceedOnlyOnEmit && !emit)
}

func toParams(values string) (map[string]string, error) {

	if values == "" {
		return map[string]string{}, nil
	}

	var params map[string]string

	result := strings.Split(values, ",")
	params = make(map[string]string)
	for _, pair := range result {
		nv := strings.Split(pair, "=")
		if len(nv) != 2 {
			return nil, fmt.Errorf("invalid settings")
		}
		params[nv[0]] = nv[1]
	}

	return params, nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"report": o.Report,
		"result": o.Result,
	}
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"value": r.Value,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {

	i.Value = values["value"]

	return nil
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Report, err = coerce.ToBool(values["report"])
	if err != nil {
		return err
	}
	o.Result = values["result"]

	return nil
}
