package uritemplates

import "testing"

func TestWithNoPlaceHolders(t *testing.T) {
	template, _ := Parse("no place holders")
	values := make(map[string]string)
	result, _ := template.ExpandString(values)
	if result != "no place holders" {
		t.Errorf("Expected %v, got %v", "no place holders", result)
	}
}

func TestLevel1(t *testing.T) {
	template, _ := Parse("{val}")
	values := make(map[string]string)
	values["val"] = "value"
	result, _ := template.ExpandString(values)
	if result != "value" {
		t.Errorf("Expected %v, got %v", "value", result)
	}
}
