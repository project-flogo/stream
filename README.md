<p align="center">
  <img src ="https://raw.githubusercontent.com/TIBCOSoftware/flogo/master/images/flogo-ecosystem_streams.png" />
</p>

<p align="center" >
  <b>Streams is a pipeline based, stream processing action for the Project Flogo Ecosystem</b>
</p>

<p align="center">
  <img src="https://travis-ci.org/TIBCOSoftware/flogo.svg"/>
  <img src="https://img.shields.io/badge/dependencies-up%20to%20date-green.svg"/>
  <img src="https://img.shields.io/badge/license-BSD%20style-blue.svg"/>
  <a href="https://gitter.im/project-flogo/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link"><img src="https://badges.gitter.im/Join%20Chat.svg"/></a>
</p>

# Flogo Stream

Edge devices have the potential for producing millions or even billions of events at rapid intervals, often times the events on their own are meaningless, hence the need to provide basic streaming operations against the slew of events.

A native streaming action as part of the Project Flogo Ecosystem accomplishes the following primary objectives:

- Enables apps to implement basic streaming constructs in a simple pipeline fashion
- Provides non-persistent state for streaming operations
  - Streams are persisted in memory until the end of the pipeline
- Serves as a pre-process pipeline for raw data to perform basic mathematical and logical operations. Ideal for feeding ML models

Some of the key highlights include:

😀 **Simple pipeline** construct enables a clean, easy way of dealing with streams of data<br/>
⏳ **Stream aggregation** across streams using time or event tumbling & sliding windows<br/>
🙌 **Join streams** from multiple event sources<br/>
🌪 **Filter** out the noise with stream filtering capabilities<br/>

## Getting Started

We’ve made building powerful streaming pipelines as easy as possible. Develop your pipelines using:

- A simple, clean JSON-based DSL
- Golang API

See the sample below of an aggregation pipeline (for brevity, the triggers and metadata of the resource has been omitted). Also don’t forget to check out the examples in the [project-flogo/stream](https://github.com/project-flogo/stream/tree/master/examples) repo.

```json
  "stages": [
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
    },
    {
      "ref": "github.com/project-flogo/contrib/activity/log",
      "input": {
        "message": "=$.result"
      }
    }
  ]
```

## Try out the example

Firstly clone the repository

Create a skeletal Flogo Application, we'll call it StreamAggregatorApp

```bash
$ flogo create -cv master StreamAggregatorApp
```

Now, install the Flogo Stream dependencies

```bash
$ cd StreamAggregatorApp/
$ flogo install github.com/project-flogo/stream@master
```

Now, install activities..

``` bash
$ flogo install github.com/project-flogo/stream/activity/aggregate
$ flogo install github.com/project-flogo/contrib/activity/log
```

Overwrite the generated flogo.json with the example...

``` bash
$ cp ../agg-flogo.json ./flogo.json
```

Fixup the flogo.json so that the name attribute is correct for the Application name we've chosen i.e replace "stream" with "StreamAggregatorApp"

```bash
$ vi flogo.json # — Change name to “StreamAggregatorApp”
```

Build it...

```bash
flogo build
```
