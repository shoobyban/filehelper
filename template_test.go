package filehelper

import "testing"

type testStruct struct {
	Template string
	Values   map[string]interface{}
	Result   string
}

func TestTemplate(t *testing.T) {
	tests := map[string]testStruct{
		"mapto": testStruct{
			Template: `{{range $key, $item := .test1}} {{ mapto $item "1:OK|2:Not OK|3:Maybe" "|:" }}{{end}}`,
			Values: map[string]interface{}{
				"test1": []string{"1", "2", "3"},
			},
			Result: " OK Not OK Maybe",
		},
		"decimal": testStruct{
			Template: `Pi: '{{ "0.812545" | decimal "6,6" }}'`,
			Result:   "Pi: '0.812545'",
		},
		"fixlen": testStruct{
			Template: `Fix: '{{ "A" | fixlen 5 }}'`,
			Result:   "Fix: 'A    '",
		},
		"fixlen2": testStruct{
			Template: `Fix: '{{ "ABCDEFG" | fixlen 5 }}'`,
			Result:   "Fix: 'ABCDE'",
		},
		"ukdate": testStruct{
			Template: `Date: '{{ "2006-01-02 15:04:05" | date "ukshort" }}'`,
			Result:   "Date: '02/01/06'",
		},
		"item": testStruct{
			Template: `C: '{{ item "1234-22" "-" 0 }}' D: '{{ item "1234-22" "-" 1 }}'`,
			Result:   "C: '1234' D: '22'",
		},
		"mapto2": testStruct{
			Template: `{{mapto "a" "a:True|b:False" "|:"}}`,
			Result:   "True",
		},
		"int": testStruct{
			Template: `'{{int "0123"}}'`,
			Result:   `'123'`,
		},
		"limit": testStruct{
			Template: `{{ limit "1234567890" 3 }}|{{ limit 1234 3 }}|{{ limit "12" 3 }}`,
			Result:   `123|123|12`,
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
