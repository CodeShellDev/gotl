package tests

import (
	"testing"

	jsonutils "github.com/codeshelldev/gotl/pkg/jsonutils"
	query "github.com/codeshelldev/gotl/pkg/query"
)

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

	expectedStr := jsonutils.Pretty(expected)
	gotStr := jsonutils.Pretty(got)

	if expectedStr != gotStr {
		t.Error("\nExpected: ", expectedStr, "\nGot: ", gotStr)
	}
}
