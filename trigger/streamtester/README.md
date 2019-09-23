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

#### Settings:
| Setting  | Required | Description |
|:---------|:---------|:------------|
| port     | true     | Port number on which control api is called 

#### Handler Settings:
| Setting        | Required | Description |
|:---------------|:---------|:------------|
| filePath       | true     | Path to a CSV file
| repeatInterval | true     | the repeat interval (1, 200 etc in millisecond), doesn't repeat if not specified
| columnNameAsKey| false    | Send as Map
| asBlock        | false    | Should the file be send as a block or stream. (Block set to true will send the csv file all at once.)

### Trigger Control API

The tester can be controlled using a REST API. 


POST /tester/start : Starts all data emission 

POST /tester/stop : Stops all data emission 

POST /tester/pause : Pauses all data emission 

POST /tester/resume : Resumes all data emission 

***Note:*** *If a handler name has been specified, the name can be use to control the data emmission for a particular handler.  ex . POST /tester/pause/myHandler*   


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
            "columnNameAsKey": true,
            "repeatInterval": "10",
            "block" : false
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
