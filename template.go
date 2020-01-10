package filehelper

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/shoobyban/mxj"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	//	"github.com/robertkrimen/otto"
)

var fmap = template.FuncMap{
	"formatUKDate":    formatUKDate,
	"limit":           limit,
	"fixlen":          fixlen,
	"fixlenr":         fixlenright,
	"sanitise":        sanitise,
	"sanitize":        sanitise,
	"last":            last,
	"reReplaceAll":    reReplaceAll,
	"replace":         replace,
	"match":           regexp.MatchString,
	"title":           strings.Title,
	"timestamp":       timestamp,
	"datetime":        datetime,
	"ukdatetime":      ukdatetime,
	"ukdate":          ukdate,
	"unixtimestamp":   unixtimestamp,
	"nanotimestamp":   nanotimestamp,
	"json":            asJSON,
	"toUpper":         strings.ToUpper,
	"upper":           strings.ToUpper,
	"toLower":         strings.ToLower,
	"lower":           strings.ToLower,
	"filter":          filterPath,
	"concat":          concat,         // concat "a" "b" => "ab"
	"empty":           empty,          // empty [] => "", ["bah"] => "bah"
	"int":             toint,          // int "0123" => 123
	"float":           tofloat,        // float "0123.234" => 123.234
	"ifthen":          conditional,    // ifthen "a" "b" => a, ifthen "" "b" => b
	"elseifthen":      notconditional, // elseifthen "a" "b" => b, elseifthen "" "b" => ""
	"mapto":           mapto,          // mapto "a" "a:True|b:False" "|:" => True
	"date":            dateFmt,        // "2017-03-31 19:59:11" |  date "06.01.02" => "17.03.31"
	"dateFrom":        dateFmtLayout,
	"decimal":         decimalFmt, // 3.1415 decimal 6,2 => 3.14
	"item":            item,       // item "a:b" ":" 0 => a
	"add":             add,
	"sub":             subtract,
	"div":             divide,
	"mul":             multiply,
	"var":             newVariable,
	"explode":         explode,
	"tojson":          jsonDecode, // backward compatibility
	"json_decode":     jsonDecode,
	"json_encode":     jsonEncode,
	"xml_decode":      xmlDecode,
	"xml_encode":      xmlEncode,
	"xml_array":       xmlArray,
	"in_array":        inArray,
	"timeformat":      timeFormat,
	"timeformatminus": timeFormatMinus,
	"unique":          unique,
	"setItem":         setItem,
	"createMap":       createMap,
	"mkSlice":         mkSlice,
	"escape":          escape,
	"seq":             seq,
	"url_path":        urlPath, // SEO, Slugify
	"urlencode":       urlencode,
	"urldecode":       urldecode,
	"md5":             md5hash,
	"json_escape":     JsonEscape,
}

var fs afero.Fs

type variable struct {
	Value interface{}
}

func (v *variable) Set(value interface{}) string {
	v.Value = value
	return ""
}

func newVariable(initialValue interface{}) *variable {
	return &variable{initialValue}
}

func md5hash(data interface{}) string {
	s := fmt.Sprintf("%#v", data)
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func urlencode(s string) string {
	return url.QueryEscape(s)
}

func urldecode(s string) (string, error) {
	r, err := url.QueryUnescape(s)
	if err != nil {
		return "", err
	}
	return r, nil
}

// From Hugo with love see https://gohugo.io/functions/seq/
// seq LAST
// seq FIRST LAST
// seq FIRST INCREMENT LAST
// It's named and used in the model of GNU's seq.
// Examples:
// anything wrong: empty list(!)
// 3: 1, 2, 3
// 1 2 4: 1, 3
// -3: -1, -2, -3
// 1 4: 1, 2, 3, 4
// 1 -2: 1, 0, -1, -2
func seq(args ...interface{}) []int {
	if len(args) < 1 || len(args) > 3 {
		return []int{}
	}

	intArgs := cast.ToIntSlice(args)
	if len(intArgs) < 1 || len(intArgs) > 3 {
		return []int{}
	}

	var inc = 1
	var last int
	var first = intArgs[0]

	if len(intArgs) == 1 {
		last = first
		if last == 0 {
			return []int{}
		} else if last > 0 {
			first = 1
		} else {
			first = -1
			inc = -1
		}
	} else if len(intArgs) == 2 {
		last = intArgs[1]
		if last < first {
			inc = -1
		}
	} else {
		inc = intArgs[1]
		last = intArgs[2]
		if inc == 0 {
			return []int{}
		}
		if first < last && inc < 0 {
			return []int{}
		}
		if first > last && inc > 0 {
			return []int{}
		}
	}

	// sanity check
	if last < -100000 {
		return []int{}
	}
	size := ((last - first) / inc) + 1

	// sanity check
	if size <= 0 || size > 2000 {
		return []int{}
	}

	seq := make([]int, size)
	val := first
	for i := 0; ; i++ {
		seq[i] = val
		val += inc
		if (inc < 0 && val < last) || (inc > 0 && val > last) {
			break
		}
	}

	return seq
}

func createMap() map[string]interface{} {
	return map[string]interface{}{}
}

func setItem(m map[string]interface{}, a string, b interface{}) map[string]interface{} {
	m[a] = b
	return m
}

func mkSlice(args ...interface{}) []interface{} {
	return args
}

func unique(e []interface{}) []interface{} {
	r := []interface{}{}

	for _, s := range e {
		if !inArray(s, r[:]) {
			r = append(r, s)
		}
	}
	return r
}

func contains(e []string, c string) bool {
	for _, s := range e {
		if s == c {
			return true
		}
	}
	return false
}

func item(s, sep string, num int) string {
	i := strings.Split(s, sep)
	if len(i) <= num {
		return ""
	}
	return i[num]
}

func escape(str string) string {
	return strings.Replace(str, "\"", "\\\"", -1)
}

func dateFmt(format, datestring string) string {
	if format == "ukshort" {
		format = "02/01/06"
	}
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, datestring)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05-0700", datestring)
		if err != nil {
			return datestring
		}
	}
	return t.Format(format)
}

func dateFmtLayout(format, datestring, layout string) string {
	if format == "ukshort" {
		format = "02/01/06"
	}
	t, err := time.Parse(layout, datestring)
	if err != nil {
		return err.Error()
	}
	return t.Format(format)
}

func decimalFmt(format string, number interface{}) string {
	num := emptyStr(number)
	f, _ := strconv.ParseFloat(num, 64)
	i := strings.Split(format, ",")
	s := fmt.Sprintf(fmt.Sprintf("%%%s.%sf", i[0], i[1]), f)
	if i[1] != "0" {
		s = strings.TrimRight(s, "0")
		if s[len(s)-1:] == "." {
			s += "0"
		}
	}
	return s
}

func formatUKDate(datestring string) string {
	return dateFmt("ukshort", datestring)
}

func limit(data interface{}, length int) interface{} {
	switch reflect.ValueOf(data).Kind() {
	case reflect.String:
		if len(data.(string)) > length {
			return data.(string)[:length]
		}
		return data

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf(fmt.Sprintf("%%-%dd", length), data)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf(fmt.Sprintf("%%-%d.4f", length), data)
	}
	return data
}

func fixlen(length int, data interface{}) interface{} {
	switch reflect.ValueOf(data).Kind() {
	case reflect.String:
		return fmt.Sprintf(fmt.Sprintf("%%-%d.%ds", length, length), data)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf(fmt.Sprintf("%%-%d.%dd", length, length), data)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf(fmt.Sprintf("%%-%d.4f", length), data)
	}
	return strings.Repeat(" ", length)
}

func fixlenright(length int, data interface{}) interface{} {
	switch reflect.ValueOf(data).Kind() {
	case reflect.String:
		return fmt.Sprintf(fmt.Sprintf("%%%d.%ds", length, length), data)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf(fmt.Sprintf("%%%d.%dd", length, length), data)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf(fmt.Sprintf("%%%d.4f", length), data)
	}
	return strings.Repeat(" ", length)
}

func urlPath(title string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return ""
	}
	prettyurl := reg.ReplaceAllString(title, "-")
	prettyurl = strings.ToLower(strings.Trim(prettyurl, "-"))
	return prettyurl
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

func datetime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func ukdate() string {
	return time.Now().Format("02/01/06")
}

func ukdatetime() string {
	return time.Now().Format("02/01/06 15:04:05")
}

func timeFormat(format string) string {
	return time.Now().Format(format)
}

func timeFormatMinus(format string, minus float64) string {
	return time.Now().Add(time.Duration(minus) * -time.Second).Format(format)
}

func unixtimestamp() int32 {
	return int32(time.Now().Unix())
}

func nanotimestamp() int64 {
	return int64(time.Now().UnixNano())
}

func empty(a interface{}) interface{} {
	k := reflect.ValueOf(a).Kind()
	if k == reflect.Int || k == reflect.Int16 || k == reflect.Int32 ||
		k == reflect.Int64 || k == reflect.Int8 || k == reflect.Bool ||
		k == reflect.Float32 || k == reflect.Float64 || k == reflect.Uint ||
		k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 ||
		k == reflect.Uint8 || k == reflect.Func {
		return a
	}
	v := reflect.ValueOf(a)
	if a == nil ||
		(k == reflect.Slice && v.Len() < 1) ||
		(k == reflect.Struct && v.NumField() < 1) ||
		(k == reflect.Map && v.Len() < 1) {
		return ""
	}
	if k == reflect.Struct {
		p := fmt.Sprintf("%#v", a)
		if p == "[]interface {}{}" || p == "map[string]interface {}{}" {
			return ""
		}
	}
	if k == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if empty(v.Index(i)) != "" {
				return a
			}
		}
		return ""
	}
	if k == reflect.Map {
		for _, mk := range v.MapKeys() {
			if empty(v.MapIndex(mk)) != "" {
				return a
			}
		}
	}
	return a
}

func emptyStr(a interface{}) string {
	k := reflect.ValueOf(a).Kind()
	if k == reflect.Int || k == reflect.Int16 || k == reflect.Int32 ||
		k == reflect.Int64 || k == reflect.Int8 || k == reflect.Bool ||
		k == reflect.Float32 || k == reflect.Float64 || k == reflect.Uint ||
		k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 ||
		k == reflect.Uint8 || k == reflect.Func {
		return fmt.Sprintf("%v", a)
	}
	if a == nil || reflect.ValueOf(a).Len() < 1 {
		return ""
	}
	return fmt.Sprintf("%v", a)
}

func asJSON(s interface{}) string {
	jsonBytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshalling %#v", s)
	}
	return string(jsonBytes)
}

func filterPath(s interface{}, p string) interface{} {
	return pathValue(strings.Split(p, "."), s, "")
}

func concat(ss ...string) string {
	return strings.Join(ss, "")
}

func toint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func tofloat(s string) (float64, error) {
	if s == "" {
		s = "0.0"
	}
	return strconv.ParseFloat(s, 64)
}

func conditional(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func notconditional(s1, s2 string) string {
	if s1 == "" {
		return s1
	}
	return s2
}

func mapto(item, mapvals, separators string) string {
	maps := strings.Split(mapvals, separators[:1])
	mapping := map[string]string{}
	for _, v := range maps {
		vv := strings.Split(v, separators[1:])
		if len(vv) < 2 {
			return ""
		}
		mapping[vv[0]] = vv[1]
	}
	if ret, ok := mapping[item]; ok {
		return ret
	}
	if ret, ok := mapping["*"]; ok {
		return ret
	}
	return item
}

// add returns the sum of a and b.
func add(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("add: unknown type for %q (%T)", av, a)
	}
}

// subtract returns the difference of b from a.
func subtract(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() - int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() - float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("subtract: unknown type for %q (%T)", av, a)
	}
}

// multiply returns the product of a and b.
func multiply(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() * int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() * float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("multiply: unknown type for %q (%T)", av, a)
	}
}

// divide returns the division of b from a.
func divide(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() / int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() / float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() / float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("divide: unknown type for %q (%T)", av, a)
	}
}

func explode(s, sep string) []string {
	return strings.Split(s, sep)
}

func inArray(needle interface{}, haystack interface{}) bool {
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(haystack)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) == true {
				return true
			}
		}
	}

	return false
}

func jsonEncode(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func xmlEncode(v interface{}) (string, error) {
	mv := mxj.Map(v.(map[string]interface{}))
	mxj.XMLEscapeChars(true)
	b, err := mv.Xml()
	return string(b), err
}

func xmlArray(v interface{}, roottag, itemtag string) (string, error) {
	mv := mxj.Map(map[string]interface{}{itemtag: v.([]interface{})})
	mxj.XMLEscapeChars(true)
	b, err := mv.XmlIndent("", "  ", roottag)
	return "<?xml version=\"1.0\"?>\n" + string(b), err
}

func decode(s, format string) (interface{}, error) {
	p := NewParser()
	res, err := p.ParseStruct(strings.NewReader(s), format)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse json '%s': %v", s, err)
	}
	return res, nil
}

func jsonDecode(s string) (interface{}, error) {
	return decode(s, "json")
}

func xmlDecode(s string) (interface{}, error) {
	return decode(s, "xml")
}

// MustTemplate parses string as Go template, using data as scope
func MustTemplate(str string, data interface{}) string {
	ret, _ := Template(str, data)
	return ret
}

// JsonEscape escapes a string for JSON
func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// TemplateDelim parses string with custom delimiters as Go template, using data as scope
func TemplateDelim(str string, data interface{}, begin, end string) (string, error) {
	tmpl, err := template.New("test").Funcs(fmap).Delims(begin, end).Parse(str)
	if err == nil {
		var doc bytes.Buffer
		err = tmpl.Execute(&doc, data)
		if err != nil {
			return "", err
		}
		return strings.Replace(doc.String(), "<no value>", "", -1), nil
	}
	return "", err
}

// Template parses string as Go template, using data as scope
func Template(str string, data interface{}) (string, error) {
	tmpl, err := template.New("test").Funcs(fmap).Parse(str)
	if err == nil {
		var doc bytes.Buffer
		err = tmpl.Execute(&doc, data)
		if err != nil {
			return "", err
		}
		return strings.Replace(doc.String(), "<no value>", "", -1), nil
	}
	return "", err
}

// RegisterFS (afero) virtual filesystem for ProcessTemplateFile
func RegisterFS(filesystem afero.Fs) {
	fs = filesystem
}

// ProcessTemplateFile processes golang template file
func ProcessTemplateFile(template string, bundle interface{}) ([]byte, error) {
	var byteValue []byte
	var err error
	if fs == nil {
		tf, err := os.Open(template)
		if err != nil {
			return nil, err
		}
		defer tf.Close()
		byteValue, err = ioutil.ReadAll(tf)
	} else {
		tf, err := fs.Open(template)
		if err != nil {
			return nil, err
		}
		byteValue, err = ioutil.ReadAll(tf)
		defer tf.Close()
	}
	if err != nil {
		return nil, err
	}
	output, err := Template(string(byteValue), bundle)
	if err != nil {
		return []byte{}, err
	}
	return []byte(output), nil
}

// MustProcessTemplateFile processes golang template file
func MustProcessTemplateFile(template string, bundle interface{}) string {
	tf, err := os.Open(template)
	if err != nil {
		return ""
	}
	byteValue, _ := ioutil.ReadAll(tf)
	output, _ := Template(string(byteValue), bundle)
	defer tf.Close()
	return output
}

// // JSTemplate parses JS code as template, using data as scope
// func JSTemplate(str string, data interface{}) string {
// 	vm := otto.New()
// 	scope, err := json.Marshal(data)
// 	script := "botl=" + string(scope) + ";" + str
// 	value, err := vm.Run(script)
// 	if err == nil {
// 		return value.String()
// 	}
// 	return ""
// }

func pathValue(keys []string, s interface{}, f string) (v interface{}) {
	var key string
	var nextkeys []string
	if len(keys) == 0 {
		if f == "" {
			return s
		}
		key = ""
		nextkeys = keys
	} else {
		key = keys[0]
		nextkeys = keys[1:]
	}
	filter := ""
	var (
		i  int64
		ok bool
	)
	var err error

	if key != "" && key[:1] == "[" && key[len(key)-1:] == "]" {
		key, filter = "", key[1:len(key)-1]
	}

	switch s.(type) {
	case map[string]interface{}:
		if key == "" {
			m := map[string]interface{}{}
			found := true
			if f != "" {
				found = false
				fparts := strings.Split(f, "=")
				for k, item := range s.(map[string]interface{}) {
					if k == fparts[0] && item == fparts[1] {
						found = true
					}
				}
			}
			if found {
				for k, item := range s.(map[string]interface{}) {
					m[k] = pathValue(nextkeys, item, filter)
				}
			}
			if len(m) > 0 {
				v = m
			}
		} else if v, ok = s.(map[string]interface{})[key]; !ok {
			err = fmt.Errorf("Key not present. [Key:%s]", key)
		}
	case []interface{}:
		array := s.([]interface{})
		a := []interface{}{}
		if f != "" {
			return a
		}
		if key == "" {
			for _, item := range array {
				pv := pathValue(nextkeys, item, filter)
				if pv != nil {
					a = append(a, pv)
				}
			}
			if len(a) == 1 {
				v = a[0]
			} else if len(a) > 0 {
				v = a
			}
		} else if i, err = strconv.ParseInt(key, 10, 64); err == nil {
			if int(i) < len(array) {
				v = array[i]
			} else {
				err = fmt.Errorf("Index out of bounds. [Index:%d] [Array:%v]", i, array)
			}
		}
	}
	return pathValue(nextkeys, v, "")
}
