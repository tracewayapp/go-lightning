package lit

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func ParseNamedQuery(driver Driver, query string, params map[string]any) (string, []any, error) {
	if driver == nil {
		return "", nil, fmt.Errorf("driver is nil")
	}

	runes := []rune(query)
	var out strings.Builder
	var args []any
	argIndex := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Single-quoted string literal: copy verbatim
		if r == '\'' {
			out.WriteRune(r)
			i++
			for i < len(runes) {
				// MySQL backslash escape: skip the next character
				if driver.SupportsBackslashEscape() && runes[i] == '\\' && i+1 < len(runes) {
					out.WriteRune(runes[i])
					i++
					out.WriteRune(runes[i])
					i++
					continue
				}
				out.WriteRune(runes[i])
				if runes[i] == '\'' {
					// Check for escaped quote ''
					if i+1 < len(runes) && runes[i+1] == '\'' {
						i++
						out.WriteRune(runes[i])
						i++
						continue
					}
					break
				}
				i++
			}
			continue
		}

		// Double-quoted string/identifier: copy verbatim
		if r == '"' {
			out.WriteRune(r)
			i++
			for i < len(runes) {
				// MySQL backslash escape: skip the next character
				if driver.SupportsBackslashEscape() && runes[i] == '\\' && i+1 < len(runes) {
					out.WriteRune(runes[i])
					i++
					out.WriteRune(runes[i])
					i++
					continue
				}
				out.WriteRune(runes[i])
				if runes[i] == '"' {
					// Check for escaped quote ""
					if i+1 < len(runes) && runes[i+1] == '"' {
						i++
						out.WriteRune(runes[i])
						i++
						continue
					}
					break
				}
				i++
			}
			continue
		}

		// Backtick identifier: copy verbatim
		if r == '`' {
			out.WriteRune(r)
			i++
			for i < len(runes) {
				out.WriteRune(runes[i])
				if runes[i] == '`' {
					// Check for escaped backtick ``
					if i+1 < len(runes) && runes[i+1] == '`' {
						i++
						out.WriteRune(runes[i])
						i++
						continue
					}
					break
				}
				i++
			}
			continue
		}

		// Colon handling
		if r == ':' {
			// Double colon :: (PG type cast) — emit literally
			if i+1 < len(runes) && runes[i+1] == ':' {
				out.WriteRune(':')
				out.WriteRune(':')
				i++
				continue
			}

			// Check if followed by a valid param start character
			if i+1 < len(runes) && isParamStart(runes[i+1]) {
				// Collect param name
				j := i + 1
				for j < len(runes) && isParamChar(runes[j]) {
					j++
				}
				name := string(runes[i+1 : j])

				val, ok := params[name]
				if !ok {
					return "", nil, fmt.Errorf("missing parameter: %s", name)
				}

				argIndex++
				args = append(args, val)

				out.WriteString(driver.Placeholder(argIndex))

				i = j - 1
				continue
			}

			// Bare colon — emit as-is
			out.WriteRune(':')
			continue
		}

		out.WriteRune(r)
	}

	return out.String(), args, nil
}

func ParseNamedQueryForModel[T any](query string, params map[string]any) (string, []any, error) {
	fieldMap, err := GetFieldMap(reflect.TypeFor[T]())
	if err != nil {
		return "", nil, err
	}
	return ParseNamedQuery(fieldMap.Driver, query, params)
}

func SelectNamed[T any](ex Executor, query string, params map[string]any) ([]*T, error) {
	parsed, args, err := ParseNamedQueryForModel[T](query, params)
	if err != nil {
		return nil, err
	}
	return Select[T](ex, parsed, args...)
}

func SelectSingleNamed[T any](ex Executor, query string, params map[string]any) (*T, error) {
	parsed, args, err := ParseNamedQueryForModel[T](query, params)
	if err != nil {
		return nil, err
	}
	return SelectSingle[T](ex, parsed, args...)
}

func UpdateNamed[T any](ex Executor, t *T, where string, params map[string]any) error {
	fieldMap, err := GetFieldMap(reflect.TypeFor[T]())
	if err != nil {
		return err
	}
	parsedWhere, args, err := ParseNamedQuery(fieldMap.Driver, where, params)
	if err != nil {
		return err
	}
	return Update[T](ex, t, parsedWhere, args...)
}

func DeleteNamed(driver Driver, ex Executor, query string, params map[string]any) error {
	parsed, args, err := ParseNamedQuery(driver, query, params)
	if err != nil {
		return err
	}
	return Delete(ex, parsed, args...)
}

func isParamStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isParamChar(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
