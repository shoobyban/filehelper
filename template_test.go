package filehelper

import "testing"

type testTemplateStruct struct {
	Template string
	Values   map[string]interface{}
	Result   string
}

func TestTemplate(t *testing.T) {
	tests := map[string]testTemplateStruct{
		"mapto": testTemplateStruct{
			Template: `{{range $key, $item := .test1}} {{ mapto $item "1:OK|2:Not OK|3:Maybe" "|:" }}{{end}}`,
			Values: map[string]interface{}{
				"test1": []string{"1", "2", "3"},
			},
			Result: " OK Not OK Maybe",
		},
		"decimal": testTemplateStruct{
			Template: `All: '{{ "0.812545" | decimal "6,6" }}, {{ "0.1" | decimal "0,1" }}, {{ decimal "0,6" .test1 }}, {{ decimal "0,6" .test2 }}, {{ decimal "0,0" .test2 }}'`,
			Values: map[string]interface{}{
				"test1": nil,
				"test2": 10010.2342342342,
			},
			Result: "All: '0.812545, 0.1, 0.0, 10010.234234, 10010'",
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
