package stringer

import (
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer *message.Printer

func init() {
	printer = message.NewPrinter(language.Vietnamese)
}

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[:1])) + string(r[1:])
}

func GroupDigit(s string) string {
	var r []rune
	for i, c := range []rune(s) {
		if i > 0 && i%3 == 0 {
			r = append(r, ',')
		}
		r = append(r, c)
	}
	return string(r)
}

type Number interface {
	int64 | uint64
}

func FormatCurrency[T Number](balance T) string {
	b := balance / 1000
	return printer.Sprintf("%d☘️", b)
}
