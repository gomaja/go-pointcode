package pointcode

import "fmt"

// NetworkIndicator is the network indicator of ITU-T Q.704 (07/96) §14.2.2:
// bits D-C of the sub-service field of the service information octet, held
// as a number from 0 to 3. It discriminates between international and national
// messages, which tells the numbering a point code belongs to (§2.4.3).
// ISPC is the numbering used with NIInternational (Q.708 (03/99) §3.4).
// Text uses the names international, international-spare, national and
// national-reserved; Q.704 defines the codes and their meanings, and these
// names are package policy.
// Check Valid before using values obtained by a direct numeric conversion.
type NetworkIndicator uint8

// Network indicator codes, bits D-C, of Q.704 (07/96) §14.2.2.
const (
	NIInternational      NetworkIndicator = 0 // 00: international network
	NIInternationalSpare NetworkIndicator = 1 // 01: spare (for international use only)
	NINational           NetworkIndicator = 2 // 10: national network
	NINationalReserved   NetworkIndicator = 3 // 11: reserved for national use
)

var networkIndicatorNames = [...]string{
	"international",
	"international-spare",
	"national",
	"national-reserved",
}

// ParseNetworkIndicator reads one of the four names exactly as String writes
// them: international, international-spare, national or national-reserved.
// Other text, including other letter case, spaces and numbers, returns zero
// and an error wrapping ErrSyntax.
func ParseNetworkIndicator(s string) (NetworkIndicator, error) {
	for code, name := range networkIndicatorNames {
		if s == name {
			return NetworkIndicator(code), nil
		}
	}
	return 0, fmt.Errorf("pointcode: parse network indicator %q: want international, international-spare, national or national-reserved: %w", s, ErrSyntax)
}

// Valid reports whether ni is one of the four two-bit codes (0 to 3).
func (ni NetworkIndicator) Valid() bool { return ni <= NINationalReserved }

// String returns the package's lowercase name for ni.
// An invalid value is shown as "invalid NetworkIndicator(n)", with n in decimal.
func (ni NetworkIndicator) String() string {
	if !ni.Valid() {
		return fmt.Sprintf("invalid NetworkIndicator(%d)", ni)
	}
	return networkIndicatorNames[ni]
}

// MarshalText returns the name, or an error wrapping ErrRange if ni is invalid.
// No text is returned on error.
func (ni NetworkIndicator) MarshalText() ([]byte, error) {
	if !ni.Valid() {
		return nil, fmt.Errorf("pointcode: marshal network indicator %d: %w", ni, ErrRange)
	}
	return []byte(networkIndicatorNames[ni]), nil
}

// UnmarshalText parses a name and replaces ni only on success.
// A nil receiver returns an error wrapping ErrSyntax.
func (ni *NetworkIndicator) UnmarshalText(text []byte) error {
	if ni == nil {
		return fmt.Errorf("pointcode: unmarshal network indicator into nil receiver: %w", ErrSyntax)
	}
	value, err := ParseNetworkIndicator(string(text))
	if err != nil {
		return err
	}
	*ni = value
	return nil
}
