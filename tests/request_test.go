package tests

import (
	"testing"

	jsonutils "github.com/codeshelldev/gotl/pkg/jsonutils"
	query "github.com/codeshelldev/gotl/pkg/query"
	templating "github.com/codeshelldev/gotl/pkg/templating"
)

func TestQueryTemplating(t *testing.T) {
	variables := map[string]interface{}{
		"value": "helloworld",
		"array": []string{
			"hello",
			"world",
		},
	}

	queryStr := "key={{.value}}&array={{.array}}"

	got, err := templating.RenderNormalizedTemplate("query", queryStr, variables)

	if err != nil {
		t.Error("Error Templating Query: ", err.Error())
	}

	expected := "key=helloworld&array=[hello,world]"

	if got != expected {
		t.Error("Expected: ", expected, "; Got: ", got)
	}
}

func TestTypedQuery(t *testing.T) {
	queryStr := "key=helloworld&array=[hello,world]&int=1"

	got, _ := query.ParseTypedQuery(queryStr)

	expected := map[string]interface{}{
		"key": "helloworld",
		"int": 1,
		"array": []string{
			"hello", "world",
		},
	}

	expectedStr := jsonutils.ToJson(expected)
	gotStr := jsonutils.ToJson(got)

	if expectedStr != gotStr {
		t.Error("\nExpected: ", expectedStr, "\nGot: ", gotStr)
	}
}
