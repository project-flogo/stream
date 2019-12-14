<!--
title: streamtester
-->
# StreamTester Trigger
This trigger gives you the ability to test your stream application using mock data provided in a csv file.

## Installation

```bash
flogo install github.com/project-flogo/stream/trigger/streamtester
```

### Configuration

### Settings:
| Name      | Type   | Description
|:---       | :---   | :---       
| port      | int    | The port for the control api to listen on - **REQUIRED**


#### Handler Settings:
| Setting        | Type    | Description |
|:---------------|:--------|:------------|
| filePath       | string  | Path to a CSV file (local file path or url) - **REQUIRED**
| emitDelay      | int     | The delay between data emission in milliseconds, the default is 100ms (min is 5ms)
| replayData     | bool    | Continuously replays the data set (default is true) 
| dataAsMap      | bool    | Convert the data to a Map, with column names as keys
| getColumnNames | bool    | Get all the column names
| allDataAtOnce  | bool    | Indicates that the data be sent all at once, otherwise one row at a time

### Output:
| Name        | Type   | Description
|:---         | :---   | :---        
| columnNames | array | The array of column names if getColumnNames was enabled
| data        | params | he data that is being emitted from the CSV (either a row or the entire set)

### Trigger Control API

The tester can be controlled using a REST API. 

| Method | Resource     | Description |
|:---    |:---          |:---         |
| POST | /tester/start  | Starts all data emission 
| POST | /tester/stop   | Stops all data emission
| POST | /tester/pause  | Pauses all data emission 
| POST | /tester/resume | Resumes all data emission 
| POST | /tester/reload | Reloads the data from all csv files

#### Fine Grained Control
If a handler name has been specified, the name can be control the data mission
for that specified handler.   

| Method | Resource     | Description |
|:---    |:---          |:---         |
| POST | /tester/start/:handlerName  | Starts data emission for the specified handler 
| POST | /tester/stop/:handlerName   | Stops data emission for the specified handler
| POST | /tester/pause/:handlerName  | Pauses data emission for the specified handler
| POST | /tester/resume/:handlerName | Resumes data emission for the specified handler
| POST | /tester/reload/:handlerName | Reloads the data from the csv for the specified handler

## Examples

### Simple
Configure the trigger to emit data from the csv file every 10 milliseconds. 

```json
{
  "triggers": [
    {
      "id": "stream-tester",
      "ref": "github.com/project-flogo/stream/trigger/streamtester",
      "handlers": [
        {
          "settings": {
            "filePath": "out.csv",
            "dataAsMap": true,
            "emitDelay": 50
          },
          "action": {
            "ref": "github.com/project-flogo/stream",
            "settings": {
              "pipelineURI": "res://stream:mystream"
            }
          }
        }
      ]
    }
  ]
}
```
