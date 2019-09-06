package support


type TelemetryService interface {
	PipelineStarted(pipelineId, instanceId string, data map[string]interface{})
	StageStarted(pipelineId, instanceId, stageId string, data map[string]interface{})
	StageFinished(pipelineId, instanceId, stageId string, data map[string]interface{})
	PipelineFinished(pipelineId, instanceId string, data map[string]interface{})
}

var telemetryService TelemetryService

func RegisterTelemetryService(service TelemetryService)  {
	telemetryService = service
}

func GetTelemetryService() TelemetryService  {
	return telemetryService
}