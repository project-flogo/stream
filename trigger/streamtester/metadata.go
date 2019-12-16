package streamtester

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Port string `md:"port,required"`
}

type HandlerSettings struct {
	FilePath       string `md:"filePath,required"`
	EmitDelay      int    `md:"emitDelay"`
	DataAsMap      bool   `md:"dataAsMap"`
	ReplayData     bool   `md:"replayData"`
	GetColumnNames bool   `md:"getColumnNames"`
	AllDataAtOnce  bool   `md:"allDataAtOnce"`
}

type Output struct {
	Data        interface{}   `md:"data"`
	ColumnNames []interface{} `md:"columnNames"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":        o.Data,
		"columnNames": o.ColumnNames,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error

	o.Data = values["data"]

	o.ColumnNames, err = coerce.ToArray(values["columnNames"])
	if err != nil {
		return err
	}

	return nil
}
