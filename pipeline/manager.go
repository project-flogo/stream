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

	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
)

const (
	uriSchemeFile = "file://"
	uriSchemeHttp = "http://"
)

//var defaultManager *Manager
//
//func GetManager() *Manager {
//	return defaultManager
//}

type Manager struct {
	pipelines map[string]*Definition

	pipelineProvider *BasicRemotePipelineProvider

	mapperFactory mapper.Factory
	resolver      resolve.CompositeResolver

	//todo switch to cache
	rfMu            sync.Mutex // protects the definition map
	remotePipelines map[string]*Definition
}

//todo fix logger
var logger = log.RootLogger()

func NewManager() *Manager {
	manager := &Manager{}
	manager.pipelines = make(map[string]*Definition)
	manager.pipelineProvider = &BasicRemotePipelineProvider{}

	return manager
}

func (m *Manager) GetPipeline(uri string) (*Definition, error) {

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

		pDef, err = NewDefinition(pConfig, m.mapperFactory, m.resolver)
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
func (*BasicRemotePipelineProvider) GetPipeline(streamURI string) (*DefinitionConfig, error) {

	var pDefBytes []byte

	if strings.HasPrefix(streamURI, uriSchemeFile) {
		// File URI
		logger.Infof("Loading Local Stream: %s\n", streamURI)
		pipelineFilePath, _ := support.URLStringToFilePath(streamURI)

		readBytes, err := ioutil.ReadFile(pipelineFilePath)
		if err != nil {
			readErr := fmt.Errorf("error reading pipeline with uri '%s', %s", streamURI, err.Error())
			logger.Errorf(readErr.Error())
			return nil, readErr
		}
		if readBytes[0] == 0x1f && readBytes[2] == 0x8b {
			pDefBytes, err = unzip(readBytes)
			if err != nil {
				decompressErr := fmt.Errorf("error uncompressing pipeline with uri '%s', %s", streamURI, err.Error())
				logger.Errorf(decompressErr.Error())
				return nil, decompressErr
			}
		} else {
			pDefBytes = readBytes

		}

	} else if strings.HasPrefix(streamURI, uriSchemeHttp) {
		// URI
		req, err := http.NewRequest("GET", streamURI, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			getErr := fmt.Errorf("error getting pipeline with uri '%s', %s", streamURI, err.Error())
			logger.Errorf(getErr.Error())
			return nil, getErr
		}
		defer resp.Body.Close()

		logger.Infof("response Status:", resp.Status)

		if resp.StatusCode >= 300 {
			//not found
			getErr := fmt.Errorf("error getting pipeline with uri '%s', status code %d", streamURI, resp.StatusCode)
			logger.Errorf(getErr.Error())
			return nil, getErr
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			readErr := fmt.Errorf("error reading pipeline response body with uri '%s', %s", streamURI, err.Error())
			logger.Errorf(readErr.Error())
			return nil, readErr
		}

		val := resp.Header.Get("flogo-compressed")
		if strings.ToLower(val) == "true" {
			decodedBytes, err := decodeAndUnzip(string(body))
			if err != nil {
				decodeErr := fmt.Errorf("error decoding compressed pipeline with uri '%s', %s", streamURI, err.Error())
				logger.Errorf(decodeErr.Error())
				return nil, decodeErr
			}
			pDefBytes = decodedBytes
		} else {
			pDefBytes = body
		}
	} else {
		return nil, fmt.Errorf("unsupport uri %s", streamURI)
	}

	var pDef *DefinitionConfig
	err := json.Unmarshal(pDefBytes, &pDef)
	if err != nil {
		logger.Errorf(err.Error())
		return nil, fmt.Errorf("error unmarshalling pipeline with uri '%s', %s", streamURI, err.Error())
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
