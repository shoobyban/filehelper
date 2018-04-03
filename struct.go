package filehelper

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/shoobyban/slog"
)

// ReadStruct reads from given file, parsing into structure
func ReadStruct(filename string) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		slog.Infof("Can't open file %s", filename)
		return nil, err
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)
	return ParseStruct(byteValue, filename)
}

// ParseStruct parses byte slice into map or slice
func ParseStruct(content []byte, filename string) (interface{}, error) {
	var out interface{}
	var err error
	if strings.HasSuffix(filename, ".xml") {
		out, err = mxj.NewMapXml(content)
		if err != nil {
			slog.Infof("Can't parse XML file %s", filename)
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".json") {
		err := json.Unmarshal(content, &out)
		if err != nil {
			slog.Infof("Can't parse JSON file %s", filename)
			return nil, err
		}
	} else {
		slog.Infof("Unknown file %s", filename)
		return nil, errors.New("Unknown file")
	}
	return out, nil
}
