package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/project-flogo/contrib/activity/log"
	"github.com/project-flogo/contrib/activity/rest"
	restTrig "github.com/project-flogo/contrib/trigger/rest"
	"github.com/project-flogo/contrib/trigger/timer"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/stream/activity/aggregate"
)

// Stores all the activities of this app
var actStream map[string]activity.Activity

func StreamTest() *api.App {
	app := api.NewApp()

	// REST Trigger to receive HTTP message
	trg := app.NewTrigger(&restTrig.Trigger{}, &restTrig.Settings{Port: 9090})
	h, _ := trg.NewHandler(&restTrig.HandlerSettings{Method: "POST", Path: "/stream"})
	h.NewAction(runActivitiesStream)

	// Timer Trigger to send HTTP message repeatedly
	tmrTrg := app.NewTrigger(&timer.Trigger{}, nil)
	tmrHandler, _ := tmrTrg.NewHandler(&timer.HandlerSettings{StartInterval: "2s", RepeatInterval: "1s"})
	tmrHandler.NewAction(runTimerActivitiesStream)

	// A REST Activity to send data to Uri
	stng := &rest.Settings{Method: "POST", Uri: "http://localhost:9090/stream",
		Headers: map[string]string{"Accept": "application/json"}}
	restAct, _ := api.NewActivity(&rest.Activity{}, stng)

	// A log Activity for logging
	logAct, _ := api.NewActivity(&log.Activity{})

	// Aggregate Activities to aggregate data obtained at 9090 port
	aggStng1 := &aggregate.Settings{Function: "accumulate", WindowType: "tumbling",
		WindowSize: 3, ProceedOnlyOnEmit: true}
	aggAct1, _ := api.NewActivity(&aggregate.Activity{}, aggStng1)

	aggStng2 := &aggregate.Settings{Function: "avg", WindowType: "tumbling",
		WindowSize: 3, ProceedOnlyOnEmit: false}
	aggAct2, _ := api.NewActivity(&aggregate.Activity{}, aggStng2)

	//Store in map to avoid activity instance recreation
	actStream = map[string]activity.Activity{}
	actStream["log"] = logAct
	actStream["rest"] = restAct
	actStream["agg1"] = aggAct1
	actStream["agg2"] = aggAct2

	return app
}

func runActivitiesStream(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {

	// Get REST Trigger Output
	trgOut := &restTrig.Output{}
	trgOut.FromMap(inputs)

	// Coerce the required outputs to string
	content, _ := coerce.ToString(trgOut.Content)

	response := handleStreamInput(content)

	reply := &restTrig.Reply{Code: 200, Data: response}
	return reply.ToMap(), nil
}

type inputStreamData struct {
	Value float64 `json:"value"`
}

func handleStreamInput(input string) map[string]interface{} {

	var in inputStreamData
	err := json.Unmarshal([]byte(input), &in)

	if err != nil {
		fmt.Println("Hello, Some problem occured during json unmarshaling")
		return nil
	}

	response := make(map[string]interface{})
	response["value"] = in.Value

	output, err := api.EvalActivity(actStream["agg1"], &aggregate.Input{Value: in.Value})

	if err != nil {
		return nil
	}

	if output["report"] == true {
		fmt.Println("[@9090]$ Accumulator Output : ", output["result"])
	}

	output, err = api.EvalActivity(actStream["agg2"], &aggregate.Input{Value: in.Value})

	if err != nil {
		return nil
	}

	if output["report"] == true {
		fmt.Printf("[@9090]$ Average Output : %0.4f\n", output["result"])
		fmt.Println()
	}

	return response
}

var num float64 = 0

func runTimerActivitiesStream(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {

	num += 0.4
	input := fmt.Sprintf("{\"value\": %f }", num)

	_, err := api.EvalActivity(actStream["rest"],
		&rest.Input{Content: input})

	if err != nil {
		return nil, err
	}

	return nil, nil
}
