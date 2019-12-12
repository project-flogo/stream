package streamtester

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Port string `md:"port"`
}

type HandlerSettings struct {
	RepeatInterval string `md:"repeatInterval,required"`
	FilePath       string `md:"filePath,required"`
	Header         bool   `md:"dataAsMap"`
	Block          bool   `md:"asBlock"`
	GetColumn      bool   `md:"getColumnNames"`
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
