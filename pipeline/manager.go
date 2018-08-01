package pipeline

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/definition"
	"github.com/TIBCOSoftware/flogo-lib/app/resource"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/flogo-lib/util"
)

const (
	uriSchemeFile    = "file://"
	uriSchemeHttp    = "http://"
	uriSchemeRes     = "res://"
	RESTYPE_PIPELINE = "pipeline"
)

var defaultManager *Manager

func GetManager() *Manager {
	return defaultManager
}

type Manager struct {
	pipelines map[string]*Definition

	pipelineProvider *BasicRemotePipelineProvider

	//todo switch to cache
	rfMu            sync.Mutex // protects the flow maps
	remotePipelines map[string]*Definition
	flowProvider    definition.Provider
}

func NewManager() *Manager {
	manager := &Manager{}
	manager.pipelines = make(map[string]*Definition)
	manager.pipelineProvider = &BasicRemotePipelineProvider{}

	return manager
}

func (m *Manager) LoadResource(config *resource.Config) error {

	var pipelineCfgBytes []byte

	if config.Compressed {
		decodedBytes, err := decodeAndUnzip(string(config.Data))
		if err != nil {
			return fmt.Errorf("error decoding compressed resource with id '%s', %s", config.ID, err.Error())
		}

		pipelineCfgBytes = decodedBytes
	} else {
		pipelineCfgBytes = config.Data
	}

	var defConfig *DefinitionConfig
	err := json.Unmarshal(pipelineCfgBytes, &defConfig)
	if err != nil {
		return fmt.Errorf("error unmarshalling pipeline resource with id '%s', %s", config.ID, err.Error())
	}

	pDef, err := NewDefinition(defConfig)
	if err != nil {
		return err
	}

	m.pipelines[config.ID] = pDef
	return nil
}

func (m *Manager) GetResource(id string) interface{} {
	return m.pipelines[id]
}

func (m *Manager) GetPipeline(uri string) (*Definition, error) {

	if strings.HasPrefix(uri, uriSchemeRes) {
		return m.pipelines[uri[6:]], nil
	}

	m.rfMu.Lock()
	defer m.rfMu.Unlock()

	if m.remotePipelines == nil {
		m.remotePipelines = make(map[string]*Definition)
	}

	pDef, exists := m.remotePipelines[uri]

	if !exists {

		pConfig, err := m.pipelineProvider.GetPipeline(uri)
		if err != nil {
			return nil, err
		}

		pDef, err = NewDefinition(pConfig)
		if err != nil {
			return nil, err
		}

		m.remotePipelines[uri] = pDef
	}

	return pDef, nil
}

type BasicRemotePipelineProvider struct {
}

//todo this can be generalized an shared with flow
func (*BasicRemotePipelineProvider) GetPipeline(pipelineURI string) (*DefinitionConfig, error) {

	var pDefBytes []byte

	if strings.HasPrefix(pipelineURI, uriSchemeFile) {
		// File URI
		logger.Infof("Loading Local Pipeline: %s\n", pipelineURI)
		flowFilePath, _ := util.URLStringToFilePath(pipelineURI)

		readBytes, err := ioutil.ReadFile(flowFilePath)
		if err != nil {
			readErr := fmt.Errorf("error reading pipeline with uri '%s', %s", pipelineURI, err.Error())
			logger.Errorf(readErr.Error())
			return nil, readErr
		}
		if readBytes[0] == 0x1f && readBytes[2] == 0x8b {
			pDefBytes, err = unzip(readBytes)
			if err != nil {
				decompressErr := fmt.Errorf("error uncompressing pipeline with uri '%s', %s", pipelineURI, err.Error())
				logger.Errorf(decompressErr.Error())
				return nil, decompressErr
			}
		} else {
			pDefBytes = readBytes

		}

	} else {
		// URI
		req, err := http.NewRequest("GET", pipelineURI, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			getErr := fmt.Errorf("error getting pipeline with uri '%s', %s", pipelineURI, err.Error())
			logger.Errorf(getErr.Error())
			return nil, getErr
		}
		defer resp.Body.Close()

		logger.Infof("response Status:", resp.Status)

		if resp.StatusCode >= 300 {
			//not found
			getErr := fmt.Errorf("error getting pipeline with uri '%s', status code %d", pipelineURI, resp.StatusCode)
			logger.Errorf(getErr.Error())
			return nil, getErr
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			readErr := fmt.Errorf("error reading pipeline response body with uri '%s', %s", pipelineURI, err.Error())
			logger.Errorf(readErr.Error())
			return nil, readErr
		}

		val := resp.Header.Get("flow-compressed")
		if strings.ToLower(val) == "true" {
			decodedBytes, err := decodeAndUnzip(string(body))
			if err != nil {
				decodeErr := fmt.Errorf("error decoding compressed pipeline with uri '%s', %s", pipelineURI, err.Error())
				logger.Errorf(decodeErr.Error())
				return nil, decodeErr
			}
			pDefBytes = decodedBytes
		} else {
			pDefBytes = body
		}
	}

	var pDef *DefinitionConfig
	err := json.Unmarshal(pDefBytes, &pDef)
	if err != nil {
		logger.Errorf(err.Error())
		return nil, fmt.Errorf("error unmarshalling pipeline with uri '%s', %s", pipelineURI, err.Error())
	}

	return pDef, nil

}

func decodeAndUnzip(encoded string) ([]byte, error) {

	decoded, _ := base64.StdEncoding.DecodeString(encoded)
	return unzip(decoded)
}

func unzip(compressed []byte) ([]byte, error) {

	buf := bytes.NewBuffer(compressed)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	jsonAsBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return jsonAsBytes, nil
}
