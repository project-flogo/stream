package streamtester

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/project-flogo/core/data/coerce"
)

type DataSet interface {
	ColumnNames() []interface{}
	DataPoint() (interface{}, bool)
	Reset()
	Reload() error
}

func NewDataSet(csvPath string, replay bool, dataAsMap, allAtOnce bool) (DataSet, error) {
	ds :=  &dataSetImpl{csvPath: csvPath, replay:replay, dataAsMap: dataAsMap, AllAtOnce: allAtOnce}
	err := ds.Reload()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

type dataSetImpl struct {
	csvPath     string
	index       int
	columnNames []interface{}
	data        []interface{}
	replay      bool
	dataAsMap   bool
	AllAtOnce   bool
	done        bool
}

func (d *dataSetImpl) Reload() error {
	d.Reset()
	csvData, err := ReadCSV(d.csvPath)
	if err != nil {
		return fmt.Errorf("unable to read csv file from '%s': %#v", d.csvPath, err)
	}

	d.columnNames, err = coerce.ToArray(csvData[0])
	if err != nil {
		return fmt.Errorf("unable to extract column names: %#v", err)
	}

	var data []interface{}

	numColumns := len(csvData[0])

	if d.dataAsMap {
		for i := 1; i < len(csvData); i++ {
			dataMap := make(map[string]interface{}, numColumns)
			for j := 0; j < numColumns; j++ {
				if num, err := strconv.ParseFloat(csvData[i][j], 64); err == nil {
					dataMap[csvData[0][j]] = num
				} else {
					dataMap[csvData[0][j]] = csvData[i][j]
				}
			}

			data = append(data, dataMap)
		}
	} else {
		for i := 1; i < len(data); i++ {
			data = append(data, csvData[i])
		}
	}

	d.data = data

	return nil
}

func (d *dataSetImpl) ColumnNames() []interface{} {
	return d.columnNames
}

func (d *dataSetImpl) DataPoint() (interface{}, bool) {

	if d.done {
		return nil, true
	}

	if d.AllAtOnce {
		if !d.replay {
			d.done = true
		}
		return d.data, d.done
	} else {
		row := d.data[d.index]
		d.index++

		if d.index == len(d.data) {

			if d.replay {
				d.index = 0
			} else {
				d.done = true
			}
		}

		return row, d.done
	}
}

func (d *dataSetImpl) Reset() {
	d.index = 0
	d.done = false
}

/////////////////////////
// CSV Util

func ReadCSV(path string) ([][]string, error) {

	if isURL(path) {
		return readCsvRemote(path)
	} else {
		return readCsvLocal(path)
	}
}

func readCsvRemote(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func readCsvLocal(path string) ([][]string, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	data, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}


func isURL(path string) bool {
	u, err := url.Parse(path)
	return err == nil && u.Scheme != "" && u.Host != ""
}