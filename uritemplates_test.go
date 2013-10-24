// Copyright 2013 Joshua Tacoma. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
			case bool:
				results[i].tests[k].expected = make([]string, 0)
			default:
				t.Errorf("Unrecognized value type %v", typ)
			}
		}
	}
	return results
}

func runSpec(t *testing.T, path string) {
	var spec = loadSpec(t, path)
	for _, group := range spec {
		for _, test := range group.tests {
			template, err := Parse(test.template)
			if err != nil {
				if len(test.expected) > 0 {
					t.Errorf("%s: %s %v", group.title, err, test.template)
				}
				continue
			}
			result, err := template.Expand(group.values)
			if err != nil {
				if len(test.expected) > 0 {
					t.Errorf("%s: %s %v", group.title, err, test.template)
				}
				continue
			} else if len(test.expected) == 0 {
				t.Errorf("%s: should have failed while parsing or expanding %v but got %v", group.title, test.template, result)
				continue
			}
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

func TestExtended(t *testing.T) {
	runSpec(t, "tests/extended-tests.json")
}

func TestNegative(t *testing.T) {
	runSpec(t, "tests/negative-tests.json")
}

func TestSpecExamples(t *testing.T) {
	runSpec(t, "tests/spec-examples.json")
}

type Location struct {
	Path    []interface{} `uri:"path"`
	Version int           `uri:"version"`
	Opts    Options       `uri:"opts"`
}

type Options struct {
	Format string `uri:"fmt"`
}

func TestExpandStruct(t *testing.T) {
	var nav = &Location{
		Path:    []interface{}{"main", "quux"},
		Version: 2,
		Opts: Options{
			Format: "pdf",
		},
	}
	if template, err := Parse("{/path*,version}{?opts*}"); err != nil {
		t.Fatalf("%v", err)
	} else if expanded, err := template.Expand(nav); err != nil {
		t.Fatalf("%v", err)
	} else if expanded != "/main/quux/2?fmt=pdf" {
		t.Fatalf(expanded)
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("http://localhost:6060{/type,path}{.fmt}{?q*}")
	}
}

func BenchmarkExpand(b *testing.B) {
	templ, _ := Parse("http://localhost:6060{/type,path}{.fmt}{?q*}")
	data := map[string]interface{}{
		"type": "pkg",
		"path": [...]string{"github.com", "jtacoma", "uritemplates"},
		"q": map[string]interface{}{
			"somequery": "x!@#$",
			"other":     "y&*()",
		},
	}
	for i := 0; i < b.N; i++ {
		templ.Expand(data)
	}
}
