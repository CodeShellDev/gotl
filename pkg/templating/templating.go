package templating

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/codeshelldev/gotl/pkg/stringutils"
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

// Create new template with funcMap
func CreateTemplateWithFunc(name string, funcMap template.FuncMap) *template.Template {
	return template.New(name).Funcs(funcMap)
}

// Render json template
func RenderJSON(data map[string]any, variables map[string]any) (map[string]any, error) {
	data, err := renderJSONTemplate(data, variables)

	if err != nil {
		return data, err
	}

	return data, nil
}

// Helper function for RenderJSON()
// recursively walks `value` and templates string values into typed values via stringutils.ToType()
// 
// `key` currently has no purpose besides errors including the key name
func RenderDataTemplateRecursively(key any, value any, variables map[string]any) (any, error) {
	var err error

	strKey := fmt.Sprintf("%v", key)

	switch asserted := value.(type) {
	case map[string]any:
		data := map[string]any{}

		for mapKey, mapValue := range asserted {
			var templatedValue any

			templatedValue, err = RenderDataTemplateRecursively(mapKey, mapValue, variables)

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

			templatedValue, err = RenderDataTemplateRecursively(arrayIndex, arrayValue, variables)

			if err != nil {
				return arrayValue, err
			}

			data = append(data, templatedValue)
		}

		return data, err

	case string:
		templt := CreateTemplateWithFunc("json:" + strKey, template.FuncMap{
			"normalize": normalize,
		})

		ApplyTemplateFunc(templt, "normalize")

		templatedValue, err := ExecuteTemplate(templt, variables)
		
		if err != nil {
			return asserted, err
		}

		return stringutils.ToType(templatedValue), nil
	default:
		return asserted, err
	}
}

func renderJSONTemplate(data map[string]any, variables map[string]any) (map[string]any, error) {
	res, err := RenderDataTemplateRecursively("", data, variables)

	mapRes, ok := res.(map[string]any)

	if !ok {
		return data, err
	}

	return mapRes, err
}

// Create template with funcMap and apply func to all variables
func CreateNormalizedTemplate(name string) *template.Template {
	templt := CreateTemplateWithFunc(name, template.FuncMap{
		"normalize": normalize,
	})

	ApplyTemplateFunc(templt, "normalize")

	return templt
}

func normalize(value any) string {
	switch asserted := value.(type) {
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