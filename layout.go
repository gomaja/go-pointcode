package pointcode

import (
	"fmt"
	"strconv"
	"strings"
)

// Layout is a national presentation layout with field widths in bits, most
// significant first. Q.704 (07/96) §2.2.1 NOTE and §2.2.6 allow national labels
// with other sub-field sizes; how such codes are written is a national
// convention, not an ITU-T one. Common examples are 8-8-8 for 24-bit codes and
// 7-4-5 or 4-3-4-3 for 16- and 14-bit codes, without attribution to a standard.
//
// ParseLayout creates an immutable value that can be copied and shared.
// The zero value is invalid: Bits returns 0, String returns "", and Format and
// Parse return errors wrapping ErrRange. A 3-8-3 layout writes "2-68-1";
// use ISPC for the Q.708 canonical form with a three-digit middle field.
type Layout struct {
	widths [32]uint8
	count  uint8
	bits   uint8
}

// ParseLayout reads widths such as "8-8-8" or "14" (one plain-decimal field).
// Each width and their sum must be 1..32. Only ASCII decimal digits and hyphen
// separators are accepted; leading zeros are allowed and String removes them.
// Syntax is checked before ranges. Failure returns the zero Layout and an
// error wrapping ErrSyntax or ErrRange.
func ParseLayout(s string) (Layout, error) {
	// Check syntax before allocating fields, including for oversized input.
	for field := range strings.SplitSeq(s, "-") {
		if !asciiDecimal(field) {
			return Layout{}, fmt.Errorf("pointcode: parse layout %q: %w", s, ErrSyntax)
		}
	}
	var layout Layout
	for field := range strings.SplitSeq(s, "-") {
		width, err := decimal(field, 32)
		if err != nil || width == 0 || uint32(layout.bits)+width > 32 {
			return Layout{}, fmt.Errorf("pointcode: parse layout %q: widths and total must be 1..32: %w", s, ErrRange)
		}
		layout.widths[layout.count] = uint8(width)
		layout.count++
		layout.bits += uint8(width)
	}
	return layout, nil
}

// Bits returns the sum of the field widths, or 0 for the invalid zero Layout.
func (l Layout) Bits() int { return int(l.bits) }

// String returns widths in canonical notation, or "" for the zero Layout.
func (l Layout) String() string {
	fields := make([]string, l.count)
	for i, width := range l.widths[:l.count] {
		fields[i] = strconv.Itoa(int(width))
	}
	return strings.Join(fields, "-")
}

// Format presents v as decimal fields without zero padding. If v needs more
// than Bits bits, it returns an error wrapping ErrRange, never a masked value.
func (l Layout) Format(v uint32) (string, error) {
	if l.bits == 0 || uint64(v) >= uint64(1)<<l.bits {
		return "", fmt.Errorf("pointcode: format %d with layout %q: %w", v, l, ErrRange)
	}
	fields := make([]string, l.count)
	remaining := l.bits
	for i, width := range l.widths[:l.count] {
		remaining -= width
		field := (v >> remaining) & fieldMask(width)
		fields[i] = strconv.FormatUint(uint64(field), 10)
	}
	return strings.Join(fields, "-"), nil
}

// Parse reads decimal ASCII fields whose count matches the layout. Leading
// zeros are allowed; signs and spaces are rejected. Every field is checked
// against its width. Syntax is checked before ranges. Failure returns zero and
// an error wrapping ErrSyntax or ErrRange; the zero Layout wraps ErrRange.
func (l Layout) Parse(s string) (uint32, error) {
	if l.bits == 0 {
		return 0, fmt.Errorf("pointcode: parse %q with zero layout: %w", s, ErrRange)
	}
	if strings.Count(s, "-") != int(l.count)-1 {
		return 0, fmt.Errorf("pointcode: parse %q with layout %q: field count: %w", s, l, ErrSyntax)
	}
	fields := strings.Split(s, "-")
	for _, field := range fields {
		if !asciiDecimal(field) {
			return 0, fmt.Errorf("pointcode: parse %q with layout %q: ASCII decimal fields required: %w", s, l, ErrSyntax)
		}
	}
	var value uint32
	for i, field := range fields {
		width := l.widths[i]
		part, err := decimal(field, fieldMask(width))
		if err != nil {
			return 0, fmt.Errorf("pointcode: parse %q with layout %q: field %d: %w", s, l, i+1, err)
		}
		value = value<<width | part
	}
	return value, nil
}

func fieldMask(width uint8) uint32 {
	return uint32(uint64(1)<<width - 1)
}

// Callers validate syntax before numeric conversion. Checking before each
// multiply/add also accepts arbitrarily many leading zeros without overflow.
func decimal(s string, maxValue uint32) (uint32, error) {
	var value uint32
	for i := range len(s) {
		digit := uint32(s[i] - '0')
		if value > maxValue/10 || (value == maxValue/10 && digit > maxValue%10) {
			return 0, fmt.Errorf("decimal field exceeds %d: %w", maxValue, ErrRange)
		}
		value = value*10 + digit
	}
	return value, nil
}
