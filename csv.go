package filehelper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
)

// WriteCSV writes headers and rows into a given file handle
func WriteCSV(file io.ReadWriter, columns []string, rows []map[string]interface{}) ([]byte, error) {
	w := csv.NewWriter(file)
	if err := w.Write(columns); err != nil {
		return nil, err
	}
	r := make([]string, len(columns))
	var ok bool
	for _, row := range rows {
		for i, column := range columns {
			if r[i], ok = row[column].(string); !ok {
				message := fmt.Sprintf("type is %T in cell for value %v", row[column], row[column])
				return nil, errors.New(message)
			}
		}
		if err := w.Write(r); err != nil {
			return nil, err
		}
	}
	w.Flush()
	byteValue, _ := ioutil.ReadAll(file)
	return byteValue, nil
}

// SplitKeys creates a map for CSV header
func SplitKeys(v interface{}) ([]string, []map[string]interface{}, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, nil, errors.New("not a map")
	}
	t := rv.Type()
	if t.Key().Kind() != reflect.String {
		return nil, nil, errors.New("not string key")
	}
	var keys []string
	values := []map[string]interface{}{v.(map[string]interface{})}
	for _, kv := range rv.MapKeys() {
		keys = append(keys, kv.String())
	}
	return keys, values, nil
}
