package pointcode

import "errors"

// ErrSyntax identifies malformed text, including non-ASCII decimal digits.
var ErrSyntax = errors.New("invalid syntax")

// ErrRange identifies a value or field width outside its supported range.
var ErrRange = errors.New("value out of range")

func asciiDecimal(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
