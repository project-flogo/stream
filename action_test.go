package stream

import (
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/engine/channels"
	"github.com/project-flogo/core/support/test"
	"github.com/project-flogo/stream/pipeline"
	"github.com/stretchr/testify/assert"
)

const testConfig string = `{
  "id": "flogo-stream",
  "ref": "github.com/project-flogo/stream",
  "settings": {
    "pipelineURI": "res://pipeline:test",
    "outputChannel": "testChan"
  }
}
`
const resData string = `{
        "metadata": {
          "input": [
            {
              "name": "input",
              "type": "integer"
            }
          ]
        },
        "stages": [
        ]
      }`

func TestActionFactory_New(t *testing.T) {

	cfg := &action.Config{}
	err := json.Unmarshal([]byte(testConfig), cfg)

	if err != nil {
		t.Error(err)
		return
	}

	af := ActionFactory{}
	ctx := test.NewActionInitCtx()

	af.Initialize(ctx)

	resourceCfg := &resource.Config{ID: "pipeline:test"}
	resourceCfg.Data = []byte(resData)
	ctx.AddResource(pipeline.RESTYPE, resourceCfg)

	channels.New("testChan", 5)

	act, err := af.New(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, act)
}
