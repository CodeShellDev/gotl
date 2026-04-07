package tests

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/codeshelldev/gotl/pkg/templating"
)

func TestTemplateFieldTransform(t *testing.T) {
	tmplStr := `Hello {{ .NAME.FIRST }} {{ .NAME.LAST }}!`

	templt, err := templating.CreateTemplateFromString("transform-field", tmplStr)

	if err != nil {
		t.Error("Error Templating:\n", err.Error())
	}

	templating.TransformTemplateFields(templt, func(fieldName string) string {
		return strings.ToLower(fieldName)
	})

	got, err := templating.ExecuteTemplate(templt, map[string]any{
		"name": map[string]any{
			"first": "John",
			"last": "Doe",
		},
	})

	if err != nil {
		t.Error("Error Templating:\n", err.Error())
	}

	expected := `Hello John Doe!`

	if got != expected {
		t.Error("Expected: ", expected, "\nGot: ", got)
	}
}

func TestTemplateApplyFunc(t *testing.T) {
	tmplStr := `Your email is {{ .base64_email }}!`

	templt, err := templating.CreateTemplateFromString("apply-func", tmplStr)

	templt.Funcs(template.FuncMap{
		"decode_base64": func (v any) string  {
			str, ok := v.(string)

			if !ok {
				return ""
			}

			decoded, _ := base64.StdEncoding.DecodeString(str)
			return string(decoded)
		},
	})

	if err != nil {
		t.Error("Error Templating:\n", err.Error())
	}

	templating.ApplyTemplateFunc(templt, "decode_base64")

	fmt.Println(templt.Root.String())

	got, err := templating.ExecuteTemplate(templt, map[string]any{
		"base64_email": "am9obi5kb2VAZXhhbXBsZS5jb20=",
	})

	if err != nil {
		t.Error("Error Templating:\n", err.Error())
	}

	expected := `Your email is john.doe@example.com!`

	if got != expected {
		t.Error("Expected: ", expected, "\nGot: ", got)
	}
}