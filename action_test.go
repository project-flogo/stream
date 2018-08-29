package stream

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/flogo-oss/core/action"
	"github.com/flogo-oss/core/app/resource"
	"github.com/flogo-oss/core/data"
	"github.com/flogo-oss/core/engine/channels"
	_ "github.com/flogo-oss/flow/test"
	"github.com/stretchr/testify/assert"
)

var testMetadata *action.Metadata

func getTestMetadata(t *testing.T) *action.Metadata {

	if testMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("action.json")
		assert.Nil(t, err)

		md := action.NewMetadata(string(jsonMetadataBytes))
		assert.NotNil(t, md)

		testMetadata = md
	}

	return testMetadata
}

const testConfig string = `{
  "id": "flogo-stream",
  "ref": "github.com/flogo-oss/stream",
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

	md := getTestMetadata(t)
	f := &ActionFactory{metadata: md}
	f.Init()

	config := &action.Config{}
	err := json.Unmarshal([]byte(testConfig), config)
	assert.Nil(t, err)

	resourceCfg := &resource.Config{ID: "pipeline:test"}
	resourceCfg.Data = []byte(resData)
	manager.LoadResource(resourceCfg)

	channels.Add("testChan:5")
	defer channels.Close()

	act, err := f.New(config)

	assert.Nil(t, err)
	assert.NotNil(t, act)
}

func TestBla(t *testing.T) {
	v, _ := data.GetResolutionDetails("$pipeline[in].input")
	fmt.Printf("value: %+v\n", v)

	v2, _ := data.GetResolutionDetails("$pipeline.input")
	fmt.Printf("value: %+v\n", v2)
}
