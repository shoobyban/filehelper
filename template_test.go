package filehelper

import "testing"

type testStruct struct {
	Template string
	Values   map[string]interface{}
	Result   string
}

func TestTemplate(t *testing.T) {
	tests := map[string]testStruct{
		"s": testStruct{
			Template: `{{range $key, $item := .test1}} {{ mapto $item "1:OK|2:Not OK|3:Maybe" "|:" }}{{end}}`,
			Values: map[string]interface{}{
				"test1": []string{"1", "2", "3"},
			},
			Result: " OK Not OK Maybe",
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
