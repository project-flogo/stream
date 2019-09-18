<!--
title: CSV Timer

-->
# Timer Trigger
This trigger provides your flogo application the ability to get data from a CSV file after a particular interval for an action 

## Installation

```bash
flogo install github.com/skothari-tibco/csvtrigger
```

## Metadata
```json
{
  "handler": {
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
| control  | false    | Enable control of trigger
| port     | false    | Port number on which control api is called (/control/resume, /control/pause, /control/restart)

#### Handler Settings:
| Setting        | Required | Description |
|:---------------|:---------|:------------|
| filePath       | true     | Path to a CSV file
| repeatInterval | false    | the repeat interval (1, 200 etc in millisecond), doesn't repeat if not specified
| id             | true     | Id of Handler
| block          | true     | Should the file be send as a block or stream. (Block set to true will send the csv file all at once.)


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
            "repeatInterval": "3000",
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
