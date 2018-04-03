package filehelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go3/text/template"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/robertkrimen/otto"
)

func formatUKDate(datestring string) string {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, datestring)
	if err != nil {
		return datestring
	}
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%d/%d", day, month, year)
}

func limit(data interface{}, length int) interface{} {
	if reflect.ValueOf(data).Kind() == reflect.String {
		return fmt.Sprintf(fmt.Sprintf("%%%ds", length), data)
	} else if reflect.ValueOf(data).Kind() == reflect.Int {
		return fmt.Sprintf(fmt.Sprintf("%%%dd", length), data)
	} else if reflect.ValueOf(data).Kind() == reflect.Float32 {
		return fmt.Sprintf(fmt.Sprintf("%%%d.4f", length), data)
	}
	return data
}

func fixlen(data interface{}, length int) interface{} {
	if reflect.ValueOf(data).Kind() == reflect.String {
		return fmt.Sprintf(fmt.Sprintf("%%%d.%ds", length, length), data)
	} else if reflect.ValueOf(data).Kind() == reflect.Int {
		return fmt.Sprintf(fmt.Sprintf("%%%d.%dd", length, length), data)
	} else if reflect.ValueOf(data).Kind() == reflect.Float32 {
		return fmt.Sprintf(fmt.Sprintf("%%%d.4f", length), data)
	}
	return data
}

func sanitise(str string) string {
	return sanitize.Name(strings.Replace(str, "/", " ", -1))
}

func last(x int, a interface{}) bool {
	return x == reflect.ValueOf(a).Len()-1
}

func reReplaceAll(pattern, repl, text string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, repl)
}

func timestamp() string {
	return time.Now().String()
}

func asJSON(s interface{}) string {
	jsonBytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshalling %#v", s)
	}
	return string(jsonBytes)
}

// Template parses string as Go template, using data as scope
func Template(str string, data interface{}) string {
	fmap := template.FuncMap{
		"formatUKDate": formatUKDate,
		"limit":        limit,
		"fixlen":       fixlen,
		"sanitise":     sanitise,
		"sanitize":     sanitise,
		"last":         last,
		"reReplaceAll": reReplaceAll,
		"match":        regexp.MatchString,
		"title":        strings.Title,
		"toUpper":      strings.ToUpper,
		"toLower":      strings.ToLower,
		"timestamp":    timestamp,
		"json":         asJSON,
		"upper":        strings.ToUpper,
		"lower":        strings.ToLower,
	}
	tmpl, err := template.New("test").Funcs(fmap).Parse(str)
	if err == nil {
		var doc bytes.Buffer
		tmpl.Execute(&doc, data)
		return strings.Replace(doc.String(), "<no value>", "", -1)
	}
	return str
}

// ProcessTemplateFile processes golang template file
func ProcessTemplateFile(template string, bundle interface{}) ([]byte, error) {
	tf, err := os.Open(template)
	if err != nil {
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(tf)
	xml := Template(string(byteValue), bundle)
	return []byte(xml), nil
}

// JSTemplate parses JS code as template, using data as scope
func JSTemplate(str string, data interface{}) string {
	vm := otto.New()
	scope, err := json.Marshal(data)
	script := "botl=" + string(scope) + ";" + str
	value, err := vm.Run(script)
	if err == nil {
		return value.String()
	}
	return ""
}
