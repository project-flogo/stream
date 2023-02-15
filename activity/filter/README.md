<!--
title: Filter
weight: 4603
-->

# Filter
This activity allows you to filter out data in a streaming pipeline, can also be used in flows.


## Installation
### Flogo CLI
```bash
flogo install github.com/project-flogo/stream/activity/filter
```

## Metadata
```json
{
  "settings": [
    {
      "name": "type",
      "type": "string",
      "required": true,
      "allowed" : ["non-zero","conditional"]
    },
    {
      "name": "proceedOnlyOnEmit",
      "type": "boolean"
    }
  ],
  "input":[
    {
      "name": "value",
      "type": "any"
    },
    {
      "name": "condition",
      "type": "boolean"
    }
  ],
  "output": [
    {
      "name": "filtered",
      "type": "boolean"
    },
    {
      "name": "value",
      "type": "any"
    }
  ]
}
```

### Details
#### Settings:
| Setting     | Required | Description |
|:------------|:---------|:------------|
| type              | true   | The type of filter to apply (ex. non-zero, conditional)
| proceedOnlyOnEmit | false  | Indicates that the next activity should proceed, true by default
_note_ : if using this activity in a flow, proceedOnlyOnEmit should be set to false

#### Input:
| Name     | Description |
|:------------|:---------|
| value    | The input value
| condition| The filter condition when filter type is 'conditional'

#### Output:
| Name     | Description |
|:------------|:---------|
| filtered    | Indicates if the value was filtered out
| value    | The input value, it is 0 if it was filtered out


## Example
The example below filters out all zero 'movement' readings

```json
{
  "id": "filter1",
  "name": "Filter",
  "activity": {
    "ref": "github.com/project-flogo/activity/filter",
    "settings": {
      "type": "non-zero"
    },
    "mappings": {
      "input": [
        {
          "type": "assign",
          "value": "movement",
          "mapTo": "value"
        }
      ]
    }
  }
}
```