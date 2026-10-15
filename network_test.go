package pointcode_test

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gomaja/go-pointcode"
)

var (
	_ encoding.TextMarshaler   = pointcode.NetworkIndicator(0)
	_ encoding.TextUnmarshaler = (*pointcode.NetworkIndicator)(nil)
)

// Q.704 (07/96) §14.2.2: literal codes pin the allocation independently of text round trips.
func TestNetworkIndicatorCodes(t *testing.T) {
	for _, tt := range []struct {
		ni   pointcode.NetworkIndicator
		bits uint8
		name string
	}{
		{pointcode.NIInternational, 0b00, "international"},
		{pointcode.NIInternationalSpare, 0b01, "international-spare"},
		{pointcode.NINational, 0b10, "national"},
		{pointcode.NINationalReserved, 0b11, "national-reserved"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.ni) != tt.bits || !tt.ni.Valid() || tt.ni.String() != tt.name {
				t.Fatalf("code %d: want bits %02b and text %q; got %q", tt.ni, tt.bits, tt.name, tt.ni.String())
			}
			text, err := tt.ni.MarshalText()
			if err != nil || string(text) != tt.name {
				t.Fatalf("marshal = %q, %v", text, err)
			}
			parsed, err := pointcode.ParseNetworkIndicator(tt.name)
			if err != nil || parsed != tt.ni {
				t.Fatalf("parse = %d, %v", parsed, err)
			}
			var decoded pointcode.NetworkIndicator
			if err := decoded.UnmarshalText(text); err != nil || decoded != tt.ni {
				t.Fatalf("unmarshal = %d, %v", decoded, err)
			}
		})
	}
}

func TestAllNetworkIndicators(t *testing.T) {
	for raw := 0; raw <= 255; raw++ {
		ni := pointcode.NetworkIndicator(raw)
		if ni.Valid() != (raw <= 3) {
			t.Fatalf("Valid(%d) = %v", raw, ni.Valid())
		}
		text, err := ni.MarshalText()
		if raw > 3 {
			if want := fmt.Sprintf("invalid NetworkIndicator(%d)", raw); ni.String() != want {
				t.Fatalf("String(%d) = %q, want %q", raw, ni.String(), want)
			}
			requireError(t, err, pointcode.ErrRange, "marshal network indicator")
			if text != nil {
				t.Fatalf("invalid %d marshaled as %q", raw, text)
			}
			continue
		}
		if err != nil || string(text) != ni.String() {
			t.Fatalf("marshal %d = %q, %v", raw, text, err)
		}
		parsed, err := pointcode.ParseNetworkIndicator(ni.String())
		if err != nil || parsed != ni {
			t.Fatalf("parse %d = %d, %v", raw, parsed, err)
		}
		decoded := pointcode.NINationalReserved
		if err := decoded.UnmarshalText(text); err != nil || decoded != ni {
			t.Fatalf("unmarshal %d = %d, %v", raw, decoded, err)
		}
	}
}

func TestNetworkIndicatorRejections(t *testing.T) {
	reject := func(t *testing.T, s string) {
		t.Helper()
		ni, err := pointcode.ParseNetworkIndicator(s)
		requireError(t, err, pointcode.ErrSyntax, "parse network indicator")
		if ni != 0 {
			t.Fatalf("error returned %d", ni)
		}
		want := fmt.Sprintf("pointcode: parse network indicator %q: want international, international-spare, national or national-reserved: invalid syntax", s)
		if err.Error() != want {
			t.Fatalf("error differs from specified syntax diagnostic")
		}
	}
	for _, s := range []string{
		"", "International", "NATIONAL", " national", "national ", "national\n", "national\x00",
		"0", "1", "2", "3", "00", "10", "11", "spare", "reserved", "national network", "international network",
		"Spare (for international use only)", "reserved for national use", "international_spare", "internationalspare",
		"national-spare", "international-reserved", "nаtional", "ｎａｔｉｏｎａｌ", "invalid NetworkIndicator(4)",
	} {
		t.Run("syntax/"+s, func(t *testing.T) { reject(t, s) })
	}
	t.Run("syntax/1MiB", func(t *testing.T) { reject(t, strings.Repeat("n", 1<<20)) })
}

func TestNetworkIndicatorTextUnmarshal(t *testing.T) {
	for _, s := range []string{"bad", "National", "2"} {
		t.Run(s, func(t *testing.T) {
			ni := pointcode.NINational
			requireError(t, ni.UnmarshalText([]byte(s)), pointcode.ErrSyntax, "parse network indicator")
			if ni != pointcode.NINational {
				t.Fatalf("failed unmarshal changed receiver to %d", ni)
			}
		})
	}
	ni := pointcode.NINational
	if err := ni.UnmarshalText([]byte("international-spare")); err != nil || ni != pointcode.NIInternationalSpare {
		t.Fatalf("successful unmarshal = %d, %v", ni, err)
	}
	var absent *pointcode.NetworkIndicator
	err := absent.UnmarshalText([]byte("national"))
	requireError(t, err, pointcode.ErrSyntax, "unmarshal network indicator into nil receiver")
}

func TestNetworkIndicatorJSON(t *testing.T) {
	value := struct{ NI pointcode.NetworkIndicator }{pointcode.NINational}
	data, err := json.Marshal(value)
	if err != nil || string(data) != `{"NI":"national"}` {
		t.Fatalf("JSON marshal = %s, %v", data, err)
	}
	var decoded struct{ NI pointcode.NetworkIndicator }
	if err := json.Unmarshal(data, &decoded); err != nil || decoded != value {
		t.Fatalf("JSON unmarshal = %v, %v", decoded, err)
	}
	err = json.Unmarshal([]byte(`{"NI":"bogus"}`), &value)
	if !errors.Is(err, pointcode.ErrSyntax) || value.NI != pointcode.NINational {
		t.Fatalf("failed JSON unmarshal = %v, %v", value, err)
	}
	if _, err := json.Marshal(pointcode.NetworkIndicator(4)); !errors.Is(err, pointcode.ErrRange) {
		t.Fatalf("invalid JSON marshal error = %v", err)
	}
}
