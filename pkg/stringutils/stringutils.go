package stringutils

import (
	"regexp"
	"strconv"
	"strings"

	jsonutils "github.com/codeshelldev/gotl/pkg/jsonutils"
)

// String to type conversion, converts typed strings into respective type
func ToType(str string) any {
	cleaned := strings.TrimSpace(str)

	//* Try JSON
	if IsEnclosedBy(cleaned, `[`, `]`) || IsEnclosedBy(cleaned, `{`, `}`) {
		data, err := jsonutils.GetJsonSafe[any](str)

		if data != nil && err == nil {
			return data
		}
	}

	//* Try String Slice
	if IsEnclosedBy(cleaned, `[`, `]`) {
		bracketsless := strings.ReplaceAll(str, "[", "")
		bracketsless = strings.ReplaceAll(bracketsless, "]", "")

		var data []string

		if Contains(str, ",") {
			data = ToArray(bracketsless)
		} else {
			data = []string{bracketsless}
		}

		if data != nil {
			if len(data) > 0 {
				return data
			}
		}
	}

	//* Try Number
	if !strings.HasPrefix(cleaned, "+") {
		intValue, intErr := strconv.Atoi(cleaned)

		if intErr == nil {
			return intValue
		}

		floatValue, floatErr := strconv.ParseFloat(cleaned, 64)

		if floatErr == nil {
			return floatValue
		}
	}

	return str
}

// Does string contain match and is match not escaped
func Contains(str string, match string) bool {
	return !IsEscaped(str, match)
}

// Is string enclosed by unescaped `char`
func IsEnclosedBy(str string, charA, charB string) bool {
	if NeedsEscapeForRegex(rune(charA[0])) {
		charA = `\` + charA
	}

	if NeedsEscapeForRegex(rune(charB[0])) {
		charB = `\` + charB
	}

	regexStr := `(^|[^\\])(\\\\)*(` + charA + `)(.*?)(^|[^\\])(\\\\)*(` + charB + ")"

	re := regexp.MustCompile(regexStr)

	matches := re.FindAllStringSubmatchIndex(str, -1)

	filtered := [][]int{}

	for _, match := range matches {
		start := match[len(match)-2]
		end := match[len(match)-1]
		char := str[start:end]

		if char != `\` {
			filtered = append(filtered, match)
		}
	}

	return len(filtered) > 0
}

// Is string completly escaped with `\`
func IsEscaped(str string, char string) bool {
	if NeedsEscapeForRegex(rune(char[0])) {
		char = `\` + char
	}

	regexStr := `(^|[^\\])(\\\\)*(` + char + ")"

	re := regexp.MustCompile(regexStr)

	matches := re.FindAllStringSubmatchIndex(str, -1)

	filtered := [][]int{}

	for _, match := range matches {
		start := match[len(match)-2]
		end := match[len(match)-1]
		char := str[start:end]

		if char != `\` {
			filtered = append(filtered, match)
		}
	}

	return len(filtered) == 0
}

// Does `char` need escaping for regex
func NeedsEscapeForRegex(char rune) bool {
	special := `.+*?()|[]{}^$\\`

	return strings.ContainsRune(special, char)
}

// Helper method for converting comma-separated string into string slices
func ToArray(sliceStr string) []string {
	if sliceStr == "" {
		return nil
	}

	rawItems := strings.Split(sliceStr, ",")
	items := make([]string, 0, len(rawItems))

	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}

	return items
}
