package templating

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
	"github.com/codeshelldev/gotl/pkg/stringutils"
)

const (
	FORMAT_FUNC = "format"
)

// Create template from string template
func CreateTemplateFromString(name string, tmplStr string) (*template.Template, error) {
	templt, err := template.New(name).Parse(tmplStr)

	if err != nil {
		return nil, err
	}

	return templt, err
}

// Parse template
func ParseTemplate(templt *template.Template, tmplStr string) error {
	_, err := templt.Parse(tmplStr)

	if err != nil {
		return err
	}

	return err
}

// Create new template and parse it
func RenderTemplate(name string, tmplStr string, variables any) (string, error) {
	templt, err := CreateTemplateFromString(name, tmplStr)

	if err != nil {
		return "", err
	}

	return ExecuteTemplate(templt, variables)
}

// Execute template
func ExecuteTemplate(templt *template.Template, variables any) (string, error) {
	var buf bytes.Buffer

	err := templt.Execute(&buf, variables)

	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Apply normalization on template, depends on SetupNormalization()
func ApplyNormalization(templt *template.Template, tmplStr string) error {
	err := ParseTemplate(templt, tmplStr)

	if err != nil {
		return err
	}

	ApplyTemplateFunc(templt, FORMAT_FUNC)

	return nil
}

// Sets up normalization before ApplyNormalization() can be called
func SetupNormalization(templt *template.Template) {
	templt.Funcs(template.FuncMap{
		FORMAT_FUNC: format,
	})
}

// Create template with normalize applied
func CreateNormalizedTemplateFromString(name string, tmplStr string) (*template.Template, error) {
	templt := template.New(name)

	SetupNormalization(templt)

	err := ApplyNormalization(templt, tmplStr)

	if err != nil {
		return templt, err
	}

	return templt, nil
}

// Template data by using go templates for string values and performing type conversion
func TemplateData(data any, variables map[string]any) (any, error) {
	templated, err := TemplateDataRecursively("", data, variables, nil)

	if err != nil {
		return data, err
	}

	return templated, nil
}

// Recursively walks `value` and templates string values into typed values via stringutils.ToType()
func TemplateDataRecursively(key string, value any, variables map[string]any, baseTemplate *template.Template) (any, error) {
	var err error

	switch asserted := value.(type) {
	case map[string]any:
		data := map[string]any{}

		for mapKey, mapValue := range asserted {
			var templatedValue any

			newKey := mapKey

			if key != "" {
				newKey = key + "." + newKey
			}

			templatedValue, err = TemplateDataRecursively(newKey, mapValue, variables, baseTemplate)

			if err != nil {
				return mapValue, err
			}

			data[mapKey] = templatedValue
		}

		return data, err

	case []any:
		data := []any{}

		for arrayIndex, arrayValue := range asserted {
			var templatedValue any

			newKey := strconv.Itoa(arrayIndex)

			if key != "" {
				newKey = key + "." + newKey
			}

			templatedValue, err = TemplateDataRecursively(newKey, arrayValue, variables, baseTemplate)

			if err != nil {
				return arrayValue, err
			}

			data = append(data, templatedValue)
		}

		return data, err

	case string:
		var templt *template.Template
		if baseTemplate != nil {
			var err error

			templt, err = baseTemplate.Clone()

			if err != nil {
				return asserted, err
			}

			templt.New(key)
		} else {
			templt = template.New(key)
		}

		SetupNormalization(templt)

		err = ApplyNormalization(templt, asserted)

		if err != nil {
			return asserted, err
		}

		templatedValue, err := ExecuteTemplate(templt, variables)
		
		if err != nil {
			return asserted, err
		}

		return stringutils.ToType(templatedValue), nil
	default:
		return asserted, err
	}
}

func format(value any) string {
	switch asserted := value.(type) {
	case map[string]any:
		return jsonutils.ToJson(asserted)
	case []string:
		return "[" + strings.Join(asserted, ",") + "]"
	case []any:
		items := make([]string, len(asserted))

		for i, item := range asserted {
			items[i] = fmt.Sprintf("%v", item)
		}

		return "[" + strings.Join(items, ",") + "]"
	default:
		return fmt.Sprintf("%v", value)
	}
}