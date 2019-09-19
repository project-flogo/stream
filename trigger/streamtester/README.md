<!--
title: streamtester

-->
# Timer Trigger
This trigger provides your flogo application the ability to get data from a CSV file after a particular interval for an action 

## Installation

```bash
flogo install github.com/project-flogo/stream/trigger/streamtester
```

## Metadata
```json
{
  "handler": {
    "name": "sample",
    "settings": [
      {
        "name": "filePath",
        "type": "string"
      },
      {
        "name": "repeatInterval",
        "type": "string"
      }
    ]
  }
}
```
### Details

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


## Example Configurations

Triggers are configured via the triggers.json of your application. The following are some example configuration of the Timer Trigger.

### Repeating
Configure the Trigger to run a flow repeating every 10 milliseconds. 

```json
{
  "triggers": [
    {
      "id": "flogo-timer",
      "ref": "github.com/project-flogo/contrib/trigger/timer",
      "handlers": [
        {
          "settings": {
            "id" : "sample",
            "filePath": "out.csv",
            "header": true,
            "repeatInterval": "10",
            "block" : false
          },
          "action": {
            "ref": "github.com/project-flogo/flow",
            "settings": {
              "flowURI": "res://flow:myflow"
            }
          }
        }
      ]
    }
  ]
}
```
### Control Example

The trigger handlers can be controlled using REST Api. 

POST /tester/resume : Will Resume all the Trigger Handlers

POST /tester/pause : Will Pause all the Trigger Handlers

POST /tester/start : Will Start all the Trigger Handlers

POST /tester/pause : Will Pause all the Trigger Handlers