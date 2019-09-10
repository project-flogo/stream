package csvtimer

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Control bool   `md:"control"`
	Port    string `md:"port"`
}

type HandlerSettings struct {
	StartInterval  string `md:"startDelay"`
	RepeatInterval string `md:"repeatInterval"`
	FilePath       string `md:"filePath"`
	Header         bool   `md:"header"`
	Count          int
	Lines          [][]string
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
