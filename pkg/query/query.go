package query

import (
	"errors"
	"strings"

	stringutils "github.com/codeshelldev/gotl/pkg/stringutils"
)

// Parse raw query (format: `a=b&c=d`) into `map[string][]string`
func ParseRawQuery(raw string) map[string][]string {
	result := make(map[string][]string)
	pairs := strings.SplitSeq(raw, "&")

	for pair := range pairs {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)

		key := parts[0]
		val := ""

		if len(parts) == 2 {
			val = parts[1]
		}

		result[key] = append(result[key], val)
	}

	return result
}

// Parse typed query values (example: `array=[a,b,c]` => `[]string{"a","b","c"}`)
func ParseTypedQuery(query string) (map[string]any, error) {
	addedData := map[string]any{}

	queryData := ParseRawQuery(query)

	if len(queryData) <= 0 {
		return nil, errors.New("query is empty")
	}

	for key, value := range queryData {
		newValue := parseTypedQueryValues(value)

		addedData[key] = newValue
	}

	return addedData, nil
}

func parseTypedQueryValues(values []string) any {
	raw := values[len(values)-1]

	return stringutils.ToType(raw)
}