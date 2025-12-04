package util

import "strings"

func ToSnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(strings.ToLower(string(r))[0]))
	}

	return string(result)
}
