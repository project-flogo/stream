package support


type TelemetryService interface {
	PipelineStarted(pipelineId string, data map[string]interface{})
	StageStarted(pipelineId, stageId string, data map[string]interface{})
	StageFinished(pipelineId, stageId string, data map[string]interface{})
	PipelineFinished(pipelineId string, data map[string]interface{})
}

var telemetryService TelemetryService

func RegisterTelemetryService(service TelemetryService)  {
	telemetryService = service
}

func GetTelemetryService() TelemetryService  {
	return telemetryService
}