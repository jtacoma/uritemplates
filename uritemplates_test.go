package uritemplates

import (
	"encoding/json"
	"os"
	"testing"
)

type spec struct {
	title  string
	values map[string]interface{}
	tests  []specTest
}
type specTest struct {
	template string
	expected []string
}

func loadSpec(t *testing.T, path string) []spec {

	file, err := os.Open(path)
	if err != nil {
		t.Errorf("Failed to load test specification: %s", err)
	}

	stat, _ := file.Stat()
	buffer := make([]byte, stat.Size())
	_, err = file.Read(buffer)
	if err != nil {
		t.Errorf("Failed to load test specification: %s", err)
	}

	var root_ interface{}
	err = json.Unmarshal(buffer, &root_)
	if err != nil {
		t.Errorf("Failed to load test specification: %s", err)
	}

	root := root_.(map[string]interface{})
	results := make([]spec, 1024)
	i := -1
	for title, spec_ := range root {
		i = i + 1
		results[i].title = title
		specMap := spec_.(map[string]interface{})
		results[i].values = specMap["variables"].(map[string]interface{})
		tests := specMap["testcases"].([]interface{})
		results[i].tests = make([]specTest, len(tests))
		for k, test_ := range tests {
			test := test_.([]interface{})
			results[i].tests[k].template = test[0].(string)
			switch typ := test[1].(type) {
			case string:
				results[i].tests[k].expected = make([]string, 1)
				results[i].tests[k].expected[0] = test[1].(string)
			case []interface{}:
				arr := test[1].([]interface{})
				results[i].tests[k].expected = make([]string, len(arr))
				for m, s := range arr {
					results[i].tests[k].expected[m] = s.(string)
				}
			default:
				t.Errorf("Unrecognized value type %v", typ)
			}
		}
	}
	return results
}

func TestStandards(t *testing.T) {
	var spec = loadSpec(t, "tests/spec-examples-by-section.json")
	for _, group := range spec {
		for _, test := range group.tests {
			template, err := Parse(test.template)
			if err != nil {
				t.Errorf("%s: %s %v", group.title, err, test.template)
			}
			result := template.ExpandString(group.values)
			pass := false
			for _, expected := range test.expected {
				if result == expected {
					pass = true
				}
			}
			if !pass {
				t.Errorf("%s: expected %v, but got %v", group.title, test.expected[0], result)
			}
		}
	}
}
