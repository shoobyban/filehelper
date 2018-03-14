package filehelper

import (
	"archive/tar"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go3/text/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"

	"github.com/clbanning/mxj"
	"github.com/robertkrimen/otto"
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

// WriteCSV writes headers and rows into a given file handle
func WriteCSV(file *os.File, columns []string, rows []map[string]interface{}) error {
	w := csv.NewWriter(file)
	if err := w.Write(columns); err != nil {
		return err
	}
	r := make([]string, len(columns))
	var ok bool
	for _, row := range rows {
		for i, column := range columns {
			if r[i], ok = row[column].(string); !ok {
				message := fmt.Sprintf("type is %T in cell for value %v", row[column], row[column])
				return errors.New(message)
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

// Template parses string as Go template, using data as scope
func Template(str string, data interface{}) string {
	fmap := template.FuncMap{
		"formatUKDate": formatUKDate,
		"limit":        limit,
		"sanitise":     sanitise,
		"sanitize":     sanitise,
		"last":         last,
		"reReplaceAll": reReplaceAll,
		"match":        regexp.MatchString,
		"title":        strings.Title,
		"toUpper":      strings.ToUpper,
		"toLower":      strings.ToLower,
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

// WriteTar will append to datafile with filename using buf data
func WriteTar(datafile, filename string, buf []byte) {
	f, err := os.OpenFile(datafile, os.O_RDWR, os.ModePerm)
	if err != nil {
		f, err = os.OpenFile(datafile, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	} else {
		fi, err := os.Stat(datafile)
		if err != nil {
			log.Fatalln(err)
		}
		if fi.Size() > 1024 {
			if _, err = f.Seek(-2<<9, os.SEEK_END); err != nil {
				log.Fatalln(err)
			}
		}
	}
	tw := tar.NewWriter(f)

	hdr := &tar.Header{
		Name:     filename,
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(buf)),
	}
	slog.Infof("Writing %s %d", filename, int64(len(buf)))
	if err := tw.WriteHeader(hdr); err != nil {
		slog.Infof("Error writing tar header %s", err.Error())
	}
	if _, err := tw.Write(buf); err != nil {
		slog.Infof("Error writing tar data %s", err.Error())
	}
	if err := tw.Close(); err != nil {
		slog.Infof("Error closing tar %s", err.Error())
	}
	f.Close()
}

// ListTar will return file list from given tar file
func ListTar(filename string) []string {
	var ret []string
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	// defer io.Copy(os.Stdout, tarReader)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg: // = regular file
			ret = append(ret, name)
		default:
			ret = append(ret, name)
		}
	}
	return ret
}

// ReadTar reads filename from given tarball and returns content
func ReadTar(tarfile, filename string) interface{} {
	f, err := os.Open(tarfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if header.Name == filename {

			bs, _ := ioutil.ReadAll(tarReader)
			return bs
		}

	}
	return nil
}

// FindInTar looks for search string in tarball, returns list of filenames and matches
func FindInTar(tarfile, search string) map[string]string {
	f, err := os.Open(tarfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	res := map[string]string{}
	tarReader := tar.NewReader(f)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bs, _ := ioutil.ReadAll(tarReader)
		if bytes.Contains(bs, []byte(search)) {
			begining := bytes.Index(bs, []byte(search))
			end := begining + len(search) + 3
			if begining > 3 {
				begining = begining - 3
			}
			if end > len(bs) {
				end = len(bs)
			}
			res[header.Name] = string(bs[begining:end])
		}
	}
	return res
}
