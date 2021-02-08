package stringer

import (
	"strings"
)

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[:1])) + string(r[1:])
}
