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

#### Handler Settings:
| Setting  | Required | Description |
|:---------|:---------|:------------|
| filePath   | true     | Path to a CSV file
| repeatInterval | false     | the repeat interval (1, 200 etc in millisecond), doesn't repeat if not specified


## Example Configurations

Triggers are configured via the triggers.json of your application. The following are some example configuration of the Timer Trigger.


### Only once immediately
Configure the Trigger to run a flow once immediately

```json
{
  "triggers": [
    {
      "id": "flogo-timer",
      "ref": "github.com/project-flogo/contrib/trigger/timer",
      "handlers": [
        {
          "settings": {
            "filePath": "path_to_file"
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
	    "filePath": "path_to_file",
            "repeatInterval": "10"
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
