package filehelper

import (
	"testing"
	"time"
)

type testTemplateStruct struct {
	Template string
	Values   map[string]interface{}
	Result   string
}

func TestTemplate(t *testing.T) {
	tests := map[string]testTemplateStruct{
		"xml_array": testTemplateStruct{
			Template: `{{xml_array .A "products" "product"}}`,
			Values:   map[string]interface{}{"A": []interface{}{map[string]interface{}{"z": 1, "p": "a"}, map[string]interface{}{"z": 2, "p": "b"}}},
			Result:   "<?xml version=\"1.0\"?>\n<products>\n  <product>\n    <p>a</p>\n    <z>1</z>\n  </product>\n  <product>\n    <p>b</p>\n    <z>2</z>\n  </product>\n</products>",
		},
		"json_escape": testTemplateStruct{
			Template: `{{json_escape .A}}`,
			Values: map[string]interface{}{"A": `dog "fish"
 cat`},
			Result: `dog \"fish\"\n cat`,
		},
		"md5": testTemplateStruct{
			Template: `{{md5 .A}}`,
			Values:   map[string]interface{}{"A": []interface{}{}, "B": 1},
			Result:   `456a37d61262ccf952ee9768cbe32d94`,
		},
		"urlencode": testTemplateStruct{
			Template: `{{urlencode "Some & % / - Query"}}`,
			Result:   `Some+%26+%25+%2F+-+Query`,
		},
		"urldecode": testTemplateStruct{
			Template: `{{urldecode "Some+%26+%25+%2F+-+Query"}}`,
			Result:   `Some & % / - Query`,
		},
		"url_path": testTemplateStruct{
			Template: `{{url_path "Some Nice - URL"}}`,
			Result:   "some-nice-url",
		},
		"intf": testTemplateStruct{
			Template: "{{if .A}}.{{end}}",
			Values:   map[string]interface{}{"A": []interface{}{}, "B": 1},
			Result:   "",
		},
		"seq0123": testTemplateStruct{
			Template: `{{range seq 0 3}}{{.}} {{ end }}`,
			Result:   `0 1 2 3 `,
		},
		"seq123": testTemplateStruct{
			Template: `{{range seq 3}}{{.}} {{ end }}`,
			Result:   `1 2 3 `,
		},
		"replace": testTemplateStruct{
			Template: `{{replace "aBcd" "B" "b"}}`,
			Result:   "abcd",
		},
		"map": testTemplateStruct{
			Template: `{{ $m := createMap }}{{ $m := setItem $m "a" "b" }}{{ $m := setItem $m "c" "d" }}{{ range $i,$item := $m }} {{ $i }}:{{ $item }}{{ end }}`,
			Result:   " a:b c:d",
		},
		"slice": testTemplateStruct{
			Template: `{{ $slice := mkSlice "a" "b" "c" }}{{ range $slice }}{{.}}{{ end }}`,
			Result:   "abc",
		},
		"unique": testTemplateStruct{
			Template: `{{ $slice := mkSlice "a" "a" "b" "b" "c" }}{{ range (unique $slice) }}{{.}}{{ end }}`,
			Result:   "abc",
		},
		"reReplaceAll": testTemplateStruct{
			Template: `{{reReplaceAll "\"" "\\\"" .A }}`,
			Values:   map[string]interface{}{"A": `ab"cd"ef`},
			Result:   `ab\"cd\"ef`,
		},
		"reReplaceAll2": testTemplateStruct{
			Template: `{{reReplaceAll "\"" "&quot;" .A }}`,
			Values:   map[string]interface{}{"A": `ab"cd"ef`},
			Result:   `ab&quot;cd&quot;ef`,
		},
		"timeformatminus": testTemplateStruct{
			Template: `{{timeformatminus "02/01/06 15:04:05" 5 }}`,
			Result:   time.Now().Add(time.Second * -5).Format("02/01/06 15:04:05"),
		},
		"timeformat": testTemplateStruct{
			Template: `{{timeformat "020106"}}`,
			Result:   time.Now().Format("020106"),
		},
		"explode": testTemplateStruct{
			Template: `{{explode "1|2|3" "|"}}`,
			Result:   "[1 2 3]",
		},
		"in_array": testTemplateStruct{
			Template: `{{in_array "1" (explode "1|2|3" "|")}}`,
			Result:   "true",
		},
		"mapto": testTemplateStruct{
			Template: `{{range $key, $item := .test1}} {{ mapto $item "1:OK|2:Not OK|3:Maybe" "|:" }}{{end}}`,
			Values: map[string]interface{}{
				"test1": []string{"1", "2", "3"},
			},
			Result: " OK Not OK Maybe",
		},
		"xmldecode": testTemplateStruct{
			Template: `{{(xml_decode .val).analysis_code_15}}`,
			Values: map[string]interface{}{
				"val": `<?xml version=\"1.0\"?><analysis_code_15>Carneval "Cool" Point</analysis_code_15>`,
			},
			Result: `Carneval "Cool" Point`,
		},
		"xmlencode": testTemplateStruct{
			Template: `{{xml_encode .}}`,
			Values: map[string]interface{}{
				"analysis_code_15": "Carneval \"Cool\" Point",
			},
			Result: `<analysis_code_15>Carneval "Cool" Point</analysis_code_15>`,
		},
		"jsondecode": testTemplateStruct{
			Template: `{{(json_decode .val).analysis_code_15}}`,
			Values: map[string]interface{}{
				"val": `{"analysis_code_15":"Carneval \"Cool\" Point"}`,
			},
			Result: `Carneval "Cool" Point`,
		},
		"jsonencode": testTemplateStruct{
			Template: `{{json_encode .}}`,
			Values: map[string]interface{}{
				"analysis_code_15": "Carneval \"Cool\" Point",
			},
			Result: `{"analysis_code_15":"Carneval \"Cool\" Point"}`,
		},
		"divdec": testTemplateStruct{
			Template: `{{$l := len .a}}Len: {{$l}}{{ $b := (div $l .b) }}{{ $a := (div $l .c) }} A+B={{ add $a $b | decimal "1,2" }}`,
			Values: map[string]interface{}{
				"a": []interface{}{1, 2, 3},
				"b": 183.33,
				"c": 149.61,
			},
			Result: `Len: 3 A+B=110.98`,
		},
		"decimal": testTemplateStruct{
			Template: `All: '{{ "0.812545" | decimal "6,6" }}, {{ "0.1" | decimal "0,1" }}, {{ decimal "0,6" .test1 }}, {{ decimal "0,6" .test2 }}, {{ decimal "0,0" .test2 }}'`,
			Values: map[string]interface{}{
				"test1": nil,
				"test2": 10010.2342342342,
			},
			Result: "All: '0.812545, 0.1, 0.0, 10010.234234, 10010'",
		},
		"dec": testTemplateStruct{
			Template: `Dec: {{ "3" | int | sub 1 }}`,
			Result:   "Dec: 2",
		},
		"decrange": testTemplateStruct{
			Template: `Dec: {{$t := var "H2G2_"}}{{ range $i, $v := .nums }}{{$t.Value}}{{ $t.Set ($v | int | sub 1) }}{{end}}`,
			Values:   map[string]interface{}{"nums": []string{"5", "3", "1"}},
			Result:   "Dec: H2G2_42",
		},
		"fixlen": testTemplateStruct{
			Template: `Fix: '{{ "A" | fixlen 5 }}'`,
			Result:   "Fix: 'A    '",
		},
		"fixlen2": testTemplateStruct{
			Template: `Fix: '{{ "ABCDEFG" | fixlen 5 }}'`,
			Result:   "Fix: 'ABCDE'",
		},
		"ukdate": testTemplateStruct{
			Template: `Date: '{{ "2006-01-02 15:04:05" | date "ukshort" }}'`,
			Result:   "Date: '02/01/06'",
		},
		"item": testTemplateStruct{
			Template: `C: '{{ item "1234-22" "-" 0 }}' D: '{{ item "1234-22" "-" 1 }}'`,
			Result:   "C: '1234' D: '22'",
		},
		"mapto2": testTemplateStruct{
			Template: `{{mapto "a" "a:True|b:False" "|:"}}`,
			Result:   "True",
		},
		"mapto3": testTemplateStruct{
			Template: `{{mapto "asdf" "a:A|b:B|*:C" "|:"}}`,
			Result:   "C",
		},
		"int": testTemplateStruct{
			Template: `'{{int "0123"}}'`,
			Result:   `'123'`,
		},
		"limit": testTemplateStruct{
			Template: `{{ limit "1234567890" 3 }}|{{ limit 1234 3 }}|{{ limit "12" 3 }}`,
			Result:   `123|1234|12`,
		},
		"empty": testTemplateStruct{
			Template: `{{ "1234567890" | empty }}|{{ "" | empty }}|{{ 3 | empty }}|{{.test1|empty}}|{{ $c := .test2|empty }}{{ concat "x" $c "y" }}|{{ $d := .test3|empty }}{{ concat "x" $d "y" }}`,
			Values: map[string]interface{}{
				"test1": []string{"1", "2", "3"},
				"test2": map[string]interface{}{},
				"test3": []interface{}{[]interface{}{}},
			},
			Result: `1234567890||3|[1 2 3]|xy|xy`,
		},
		"filter": testTemplateStruct{
			Template: `{{ $c := filter .countries "data.[iso=GB]" }}{{ $c.name }}`,
			Values: map[string]interface{}{
				"countries": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"iso":  "GB",
							"name": "Great Britain",
						},
						map[string]interface{}{
							"iso":  "US",
							"name": "United States",
						},
					},
				},
			},
			Result: `Great Britain`,
		},
		"deepfilter": testTemplateStruct{
			Template: `{{ $c := filter .countries "data.[iso=GB]" }}{{ $r := filter $c "states.[name=Surrey]" }}{{ $r.id }}`,
			Values: map[string]interface{}{
				"countries": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"iso":  "GB",
							"name": "Great Britain",
							"states": []interface{}{
								map[string]interface{}{
									"id":   1,
									"name": "Surrey",
								},
							},
						},
						map[string]interface{}{
							"iso":  "US",
							"name": "United States",
							"states": []interface{}{
								map[string]interface{}{
									"id":   2,
									"name": "Alabama",
								},
							},
						},
					},
				},
			},
			Result: `1`,
		},
	}
	for name, test := range tests {
		res, err := Template(test.Template, test.Values)
		if err != nil {
			panic(err)
		}
		if res != test.Result {
			t.Errorf("%s: %#v != %#v", name, res, test.Result)
		}
	}
}
