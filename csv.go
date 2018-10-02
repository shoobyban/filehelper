package filehelper

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	csvmap "github.com/recursionpharma/go-csv-map"
	"github.com/shoobyban/slog"
)

// WriteCSV writes headers and rows into a given file handle and reads it back as []byte
func WriteCSV(file *os.File, columns []string, rows []map[string]interface{}) ([]byte, error) {
	w := csv.NewWriter(file)
	err := OnlyWriteCSV(*w, columns, rows)
	if err != nil {
		return []byte{}, err
	}
	byteValue, _ := ioutil.ReadAll(file)
	return byteValue, nil
}

// ReadCSV reads csv into []map[string]string
func ReadCSV(filename string) ([]map[string]string, error) {
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	r := csvmap.NewReader(bufio.NewReader(csvFile))
	r.Columns, err = r.ReadHeader()
	if err != nil {
		slog.Errorf("Error reading csv header %v", err)
	}
	return r.ReadAll()
}

// OnlyWriteCSV writes headers and rows into a given file handle
func OnlyWriteCSV(w csv.Writer, columns []string, rows []map[string]interface{}) error {
	if err := w.Write(columns); err != nil {
		return err
	}
	r := make([]string, len(columns))
	var ok bool
	for _, row := range rows {
		for i, column := range columns {
			if r[i], ok = row[column].(string); !ok {
				message := fmt.Sprintf("type is %T in cell for value %v", row[column], row[column])
				return fmt.Errorf(message)
			}
		}
		if err := w.Write(r); err != nil {
			return err
		}
	}
	w.Flush()
	return nil
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
