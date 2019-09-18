package testTimer

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Control bool   `md:"control"`
	Port    string `md:"port"`
}

type HandlerSettings struct {
	Id 			   string `md:"id,required"`
	StartInterval  string `md:"startDelay"`
	RepeatInterval string `md:"repeatInterval,required"`
	FilePath       string `md:"filePath"`
	Header         bool   `md:"header"`
	Block          bool   `md:"block,required"`
	Count          int
	Lines          [][]string
	Ch chan int
}

type Output struct {
	Data  interface{} `md:"data"`
	Error interface{} `md:"error"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":  o.Data,
		"error": o.Error,
	}
}
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error

	o.Data = values["data"]

	o.Error, err = coerce.ToAny(values["error"])
	if err != nil {
		return err
	}
	return nil

}
