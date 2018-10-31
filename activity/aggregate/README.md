<!--
title: Aggregate
weight: 4603
-->

# Aggregate
This activity allows you to aggregate data and calculate an average or sliding average.


## Installation
### Flogo Web
This activity comes out of the box with the Flogo Web UI
### Flogo CLI
```bash
flogo install github.com/project-flogo/stream/activity/aggregate
```

## Schema

```json
{
  "settings": [
    {
      "name": "function",
      "type": "string",
      "required": true,
      "allowed" : ["avg", "sum", "min", "max", "count", "accumulate"]
    },
    {
      "name": "windowType",
      "type": "string",
      "required": true,
      "allowed" : ["tumbling", "sliding", "timeTumbling", "timeSliding"]
    },
    {
      "name": "windowSize",
      "type": "integer",
      "required": true
    },
    {
      "name": "resolution",
      "type": "integer"
    },
    {
      "name": "proceedOnlyOnEmit",
      "type": "boolean"
    },
    {
      "name": "additionalSettings",
      "type": "string"
    }
  ],
  "input":[
    {
      "name": "value",
      "type": "any"
    }
  ],
  "output": [
    {
      "name": "result",
      "type": "any"
    },
    {
      "name": "report",
      "type": "boolean"
    }
  ]
}
```

### Details
#### Settings:
| Setting     | Required | Description |
|:------------|:---------|:------------|
| function    | true     | The aggregate function (ex. avg,sum,min,max,count)|
| windowType  | true     | The type of window (ex. tumbling,sliding,timeTumbling,timeSliding)|
| windowSize  | true     | The window size of the values to aggregate |
| resolution        | false    | The window resolution |
| proceedOnlyOnEmit | false    | Proceed to the next activity only on emit of a result |
| additionalSettings| false    | Additional settings for particular functions |
_note_ : if using this activity in a flow, proceedOnlyOnEmit should be set to false

#### Input:
| Name     | Description |
|:------------|:---------|
| value    | The input value

#### Output:
| Name     | Description |
|:------------|:---------|
| report    | Indicates if the value should be reported
| result    | The result of the aggregation


## Example
The below example sums of all the values in a tumbling time window of 5 seconds:

```json
{
  "ref": "github.com/project-flogo/stream/activity/aggregate",
  "settings": {
    "function": "sum",
    "windowType": "timeTumbling",
    "windowSize": "5000"
  },
  "input": {
    "value": "=$.input"
  }
}
```