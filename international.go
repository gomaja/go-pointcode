package pointcode

import (
	"fmt"
	"strings"
)

// ISPC is an international signalling point code: one 14-bit pure binary value
// (Q.704 (07/96) §2.2.3, Figure 3; Q.708 (03/99) §§3.4, 5.1, Figure 1).
// It is used in messages whose network indicator is NIInternational
// (Q.708 §§1, 3.4).
// Its fields are for administration and display; protocol and signalling
// functions must treat it as a single value (Q.708 §5.2).
// SANC ranges 0 and 1 are reserved but may be released (Q.708 §§7.4, 13.4–13.5),
// so this type accepts them. Validity describes the format, not an assignment.
// Check Valid before using values obtained by a direct numeric conversion.
type ISPC uint16

// MaxISPC is the largest 14-bit international signalling point code.
const MaxISPC ISPC = 1<<14 - 1

// NewISPC combines the administrative fields in Q.708 (03/99) §5.1, Figure 1.
// Zone and signallingPoint must be 0..7; areaNetwork spans 0..255.
func NewISPC(zone, areaNetwork, signallingPoint uint8) (ISPC, error) {
	if zone > 7 || signallingPoint > 7 {
		return 0, fmt.Errorf("pointcode: new ISPC (%d, %d, %d): %w", zone, areaNetwork, signallingPoint, ErrRange)
	}
	return ISPC(zone)<<11 | ISPC(areaNetwork)<<3 | ISPC(signallingPoint), nil
}

// ParseISPC reads Q.708 (03/99) §5.1 notation, such as "2-068-1".
// First and last fields must each be one ASCII digit; the middle field may
// contain one to three ASCII digits. Spaces, signs and other forms are rejected.
// Syntax is checked before ranges; a failure returns zero and wraps ErrSyntax
// or ErrRange. String produces the canonical three-digit middle field.
func ParseISPC(s string) (ISPC, error) {
	fields, err := parseInternational(s, 3)
	if err != nil {
		return 0, fmt.Errorf("pointcode: parse ISPC %q: %w", s, err)
	}
	return ISPC(fields[0])<<11 | ISPC(fields[1])<<3 | ISPC(fields[2]), nil
}

// Valid reports whether p fits the 14-bit format; it does not check assignment.
func (p ISPC) Valid() bool { return p <= MaxISPC }

// Zone returns bits NML, identifying the SANC range of Q.708 (03/99) §§7.4, 13,
// commonly called the zone. This field has only administrative use (§5.2).
func (p ISPC) Zone() uint8 { return uint8(p>>11) & 7 }

// AreaNetwork returns bits K-D of Q.708 (03/99) §5.1, Figure 1.
func (p ISPC) AreaNetwork() uint8 { return uint8(p >> 3) }

// SignallingPoint returns the signalling point identification, bits CBA
// of Q.708 (03/99) §5.1, Figure 1.
func (p ISPC) SignallingPoint() uint8 { return uint8(p & 7) }

// SANC returns the signalling area/network code (Q.708 (03/99) §5.1).
func (p ISPC) SANC() SANC { return SANC(p >> 3) }

// String returns canonical Q.708 (03/99) §5.1 notation ("2-068-1").
// An invalid value is shown as "invalid ISPC(n)", with n in decimal.
func (p ISPC) String() string {
	if !p.Valid() {
		return fmt.Sprintf("invalid ISPC(%d)", p)
	}
	return fmt.Sprintf("%d-%03d-%d", p.Zone(), p.AreaNetwork(), p.SignallingPoint())
}

// MarshalText returns canonical notation, or an error wrapping ErrRange if
// p is invalid. No text is returned on error.
func (p ISPC) MarshalText() ([]byte, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("pointcode: marshal ISPC %d: %w", p, ErrRange)
	}
	return []byte(p.String()), nil
}

// UnmarshalText parses notation and replaces p only on success.
// A nil receiver returns an error wrapping ErrSyntax.
func (p *ISPC) UnmarshalText(text []byte) error {
	if p == nil {
		return fmt.Errorf("pointcode: unmarshal ISPC into nil receiver: %w", ErrSyntax)
	}
	value, err := ParseISPC(string(text))
	if err != nil {
		return err
	}
	*p = value
	return nil
}

// SANC is a signalling area/network code: the 11 most significant bits of
// an ISPC (Q.708 (03/99) §5.1, Figure 1). Its fields have administrative use
// only (§5.2). Ranges 0 and 1 are accepted (§§7.4, 13.4–13.5).
// Check Valid before using values obtained by a direct numeric conversion.
type SANC uint16

// MaxSANC is the largest 11-bit signalling area/network code.
const MaxSANC SANC = 1<<11 - 1

// NewSANC combines zone (0..7) and areaNetwork (0..255), per Q.708 (03/99) §5.1.
func NewSANC(zone, areaNetwork uint8) (SANC, error) {
	if zone > 7 {
		return 0, fmt.Errorf("pointcode: new SANC (%d, %d): %w", zone, areaNetwork, ErrRange)
	}
	return SANC(zone)<<8 | SANC(areaNetwork), nil
}

// ParseSANC reads the first two fields of Q.708 (03/99) §5.1 notation,
// such as "2-068". The first field is one ASCII digit; the second is one to
// three ASCII digits. Syntax is checked before ranges; failures return zero
// and wrap ErrSyntax or ErrRange.
func ParseSANC(s string) (SANC, error) {
	fields, err := parseInternational(s, 2)
	if err != nil {
		return 0, fmt.Errorf("pointcode: parse SANC %q: %w", s, err)
	}
	return SANC(fields[0])<<8 | SANC(fields[1]), nil
}

// Valid reports whether s fits the 11-bit format; it does not check assignment.
func (s SANC) Valid() bool { return s <= MaxSANC }

// Zone returns bits NML, the SANC range commonly called the zone
// (Q.708 (03/99) §§5.1, 7.4, 13). It has administrative use only (§5.2).
func (s SANC) Zone() uint8 { return uint8(s>>8) & 7 }

// AreaNetwork returns bits K-D of Q.708 (03/99) §5.1, Figure 1.
func (s SANC) AreaNetwork() uint8 { return uint8(s) }

// ISPC combines s with one of eight signalling point identifications (0..7),
// per Q.708 (03/99) §7.2. An invalid SANC or identification wraps ErrRange.
func (s SANC) ISPC(signallingPoint uint8) (ISPC, error) {
	if !s.Valid() || signallingPoint > 7 {
		return 0, fmt.Errorf("pointcode: SANC %d to ISPC with point %d: %w", s, signallingPoint, ErrRange)
	}
	return ISPC(s)<<3 | ISPC(signallingPoint), nil
}

// String returns canonical SANC notation ("2-068"), per Q.708 (03/99) §5.1.
// An invalid value is shown as "invalid SANC(n)", with n in decimal.
func (s SANC) String() string {
	if !s.Valid() {
		return fmt.Sprintf("invalid SANC(%d)", s)
	}
	return fmt.Sprintf("%d-%03d", s.Zone(), s.AreaNetwork())
}

// MarshalText returns canonical notation, or an error wrapping ErrRange if
// s is invalid. No text is returned on error.
func (s SANC) MarshalText() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("pointcode: marshal SANC %d: %w", s, ErrRange)
	}
	return []byte(s.String()), nil
}

// UnmarshalText parses notation and replaces s only on success.
// A nil receiver returns an error wrapping ErrSyntax.
func (s *SANC) UnmarshalText(text []byte) error {
	if s == nil {
		return fmt.Errorf("pointcode: unmarshal SANC into nil receiver: %w", ErrSyntax)
	}
	value, err := ParseSANC(string(text))
	if err != nil {
		return err
	}
	*s = value
	return nil
}

func parseInternational(s string, count int) ([3]uint8, error) {
	var result [3]uint8
	if strings.Count(s, "-") != count-1 {
		return result, fmt.Errorf("expected %d decimal fields: %w", count, ErrSyntax)
	}
	fields := strings.Split(s, "-")
	for i, field := range fields {
		maxDigits := 1
		if i == 1 {
			maxDigits = 3
		}
		if len(field) > maxDigits || !asciiDecimal(field) {
			return result, fmt.Errorf("field %d must contain 1..%d ASCII digits: %w", i+1, maxDigits, ErrSyntax)
		}
	}
	for i, field := range fields {
		var value uint16
		for j := range len(field) {
			value = value*10 + uint16(field[j]-'0')
		}
		maxValue := uint16(7)
		if i == 1 {
			maxValue = 255
		}
		if value > maxValue {
			return [3]uint8{}, fmt.Errorf("field %d exceeds %d: %w", i+1, maxValue, ErrRange)
		}
		result[i] = uint8(value)
	}
	return result, nil
}
