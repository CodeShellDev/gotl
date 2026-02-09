package stringutils

import (
	"strconv"
	"strings"

	jsonutils "github.com/codeshelldev/gotl/pkg/jsonutils"
)

// String to type conversion, converts typed strings into respective type
func ToType(str string) any {
	cleaned := strings.TrimSpace(str)

	// try json
	if IsEnclosedByAndUnescaped(cleaned, '[', ']') || IsEnclosedByAndUnescaped(cleaned, '{', '}') {
		data, err := jsonutils.GetJsonSafe[any](str)

		if data != nil && err == nil {
			return data
		}
	}

	// try string slice
	if IsEnclosedByAndUnescaped(cleaned, '[', ']') {
		bracketsless := strings.ReplaceAll(str, "[", "")
		bracketsless = strings.ReplaceAll(bracketsless, "]", "")

		var data []string

		if ContainsRune(str, ',') {
			data = ToArray(bracketsless)
		} else {
			unescaped := UnescapeAll(bracketsless)
			data = []string{unescaped}
		}

		if data != nil {
			if len(data) > 0 {
				return data
			}
		}
	}

	// try number
	if !strings.HasPrefix(cleaned, "+") {
		// number is not literal
		if !IsRuneEscaped(cleaned, 0) {
			unescaped := UnescapeAll(cleaned)

			intValue, intErr := strconv.Atoi(unescaped)

			if intErr == nil {
				return intValue
			}

			floatValue, floatErr := strconv.ParseFloat(unescaped, 64)

			if floatErr == nil {
				return floatValue
			}
		}
	}

	return str
}

// Removes single backslash escapes from the entire string (`\a` => `a`, `\\a` => `\a`)
func UnescapeAll(str string) string {
    runes := []rune(str)
    result := []rune{}
    i := 0

    for i < len(runes) {
        if runes[i] == '\\' {
            if i + 1 < len(runes) {
                next := runes[i + 1]
				// check if single(!) backslash
                if i == 0 || runes[i - 1] != '\\' {
					// single backslash => remove it
                    result = append(result, next)

                    i += 2

                    continue
                } else {
					// double backslash => keep one and remove the other
                    result = append(result, '\\', next)

                    i += 2

                    continue
                }
            } else {
				// trailing backslash => keep

                result = append(result, '\\')
                i++
                continue
            }
        } else {
            result = append(result, runes[i])
            i++
        }
    }

    return string(result)
}

// Removes the escaping backslash for a specific rune in the string
func UnescapeRune(str string, target rune) string {
    runes := []rune(str)
    result := []rune{}

    i := 0
    for i < len(runes) {
        r := runes[i]

        if r == '\\' && i + 1 < len(runes) && runes[i + 1] == target {
			// single backslash => remove it
            result = append(result, runes[i + 1])

            i += 2
        } else if r == '\\' && i + 2 < len(runes) && runes[i + 1] == '\\' && runes[i + 2] == target {
			// double backslash => keep one and remove the other
            result = append(result, '\\', runes[i + 2])

            i += 3
        } else {
            result = append(result, r)
            i++
        }
    }

    return string(result)
}

// Does string contain match and is match not escaped
func ContainsRune(str string, match rune) bool {
	return !IsEscaped(str, match)
}

// Checks if str starts with charA and ends with charB (and are unescaped)
func IsEnclosedByAndUnescaped(str string, charA, charB rune) bool {
    runes := []rune(str)
    if len(runes) < 2 {
        return false
    }

    if runes[0] != charA || IsRuneEscaped(str, 0) {
        return false
    }

    lastIndex := len(runes) - 1
    if runes[lastIndex] != charB || IsRuneEscaped(str, lastIndex) {
        return false
    }

    return true
}

// Checks if the rune at index `pos` in `str` is escaped by a single backslash
func IsRuneEscaped(str string, pos int) bool {
    runes := []rune(str)
    if pos <= 0 || pos >= len(runes) {
        return false
    }

    // Only consider the immediately preceding rune
    return runes[pos - 1] == '\\' && (pos - 2 < 0 || runes[pos - 2] != '\\')
}

// Checks if every occurrence of `char` in `str` is escaped by `\`
func IsEscaped(str string, char rune) bool {
    runes := []rune(str)

    for i, r := range runes {
        if r == char && !IsRuneEscaped(str, i) {
            return false
        }
    }

    return true
}

// Checks if str starts with target rune and it is not escaped by `\`
func HasUnescapedPrefix(str string, target rune) bool {
    runes := []rune(str)

    if len(runes) == 0 {
        return false
    }

    if runes[0] != target {
        return false
    }

    return !IsRuneEscaped(str, 0)
}

// Checks if `char` needs escaping for regex
func NeedsEscapeForRegex(char rune) bool {
	special := `.+*?()|[]{}^$\\`

	return strings.ContainsRune(special, char)
}

// Helper method for converting a (!unescaped) comma-separated string into string slices
func ToArray(sliceStr string) []string {
    if sliceStr == "" {
        return nil
    }

    runes := []rune(sliceStr)
    var items []string
    var current []rune

    for i := 0; i < len(runes); i++ {
        r := runes[i]

        if r == ',' && !IsRuneEscaped(sliceStr, i) {
            // unescaped comma => end of current item

            item := UnescapeAll(string(current))
            item = strings.TrimSpace(item)

            if item != "" {
                items = append(items, item)
            }

            current = []rune{}
        } else {
            current = append(current, r)
        }
    }

    // add last item
    if len(current) > 0 {
        item := UnescapeAll(string(current))
        item = strings.TrimSpace(item)

        if item != "" {
            items = append(items, item)
        }
    }

    return items
}