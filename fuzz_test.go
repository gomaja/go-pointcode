package pointcode_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/gomaja/go-pointcode"
)

func checkFuzzError(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, pointcode.ErrSyntax) && !errors.Is(err, pointcode.ErrRange) {
		t.Fatalf("unclassified error: %v", err)
	}
}

func FuzzParseISPC(f *testing.F) {
	for _, s := range []string{"2-068-1", "2-68-1", "0-000-0", "7-255-7", "1-000-0", "", "8-256-9", "2-0068-1", "2-٠٦٨-1", "2--1"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		p, err := pointcode.ParseISPC(s)
		if err != nil {
			checkFuzzError(t, err)
			if p != 0 {
				t.Fatal("error returned nonzero ISPC")
			}
			return
		}
		if !p.Valid() {
			t.Fatalf("accepted invalid ISPC %d", p)
		}
		canonical := p.String()
		parsed, err := pointcode.ParseISPC(canonical)
		if err != nil || parsed != p || parsed.String() != canonical {
			t.Fatalf("unstable canonical ISPC %q: %d, %v", canonical, parsed, err)
		}
		text, err := p.MarshalText()
		if err != nil || string(text) != canonical {
			t.Fatalf("marshal differs: %q, %v", text, err)
		}
		var decoded pointcode.ISPC
		if err := decoded.UnmarshalText(text); err != nil || decoded != p {
			t.Fatalf("unmarshal differs: %d, %v", decoded, err)
		}
	})
}

func FuzzParseSANC(f *testing.F) {
	for _, s := range []string{"2-068", "2-68", "0-000", "7-255", "1-000", "", "8-256", "2-0068", "2-٠٦٨", "2-"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		code, err := pointcode.ParseSANC(s)
		if err != nil {
			checkFuzzError(t, err)
			if code != 0 {
				t.Fatal("error returned nonzero SANC")
			}
			return
		}
		if !code.Valid() {
			t.Fatalf("accepted invalid SANC %d", code)
		}
		canonical := code.String()
		parsed, err := pointcode.ParseSANC(canonical)
		if err != nil || parsed != code || parsed.String() != canonical {
			t.Fatalf("unstable canonical SANC %q: %d, %v", canonical, parsed, err)
		}
		text, err := code.MarshalText()
		if err != nil || string(text) != canonical {
			t.Fatalf("marshal differs: %q, %v", text, err)
		}
		var decoded pointcode.SANC
		if err := decoded.UnmarshalText(text); err != nil || decoded != code {
			t.Fatalf("unmarshal differs: %d, %v", decoded, err)
		}
	})
}

func FuzzParseLayout(f *testing.F) {
	for _, s := range []string{"3-8-3", "8-8-8", "7-4-5", "4-3-4-3", "14", "32", "1", "08-008-8", "", "0", "33", "16-17", "8--8", "+8", "８"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		layout, err := pointcode.ParseLayout(s)
		if err != nil {
			checkFuzzError(t, err)
			if layout.Bits() != 0 || layout.String() != "" {
				t.Fatal("error returned nonzero layout")
			}
			return
		}
		// Independently sum decimal widths; a stable round trip alone would
		// miss a parser that consistently discards or changes a field.
		bits := 0
		for _, field := range strings.Split(s, "-") {
			width, err := strconv.ParseUint(field, 10, 8)
			if err != nil || width < 1 || width > 32 {
				t.Fatalf("accepted invalid width %q", field)
			}
			bits += int(width)
		}
		if bits < 1 || bits > 32 || layout.Bits() != bits {
			t.Fatalf("invalid bit count %d (want %d)", layout.Bits(), bits)
		}
		canonical := layout.String()
		parsed, err := pointcode.ParseLayout(canonical)
		if err != nil || parsed.Bits() != bits || parsed.String() != canonical {
			t.Fatalf("unstable layout %q: %q, %v", canonical, parsed.String(), err)
		}
	})
}

func FuzzLayoutParseFormat(f *testing.F) {
	for _, seed := range []struct {
		layout, text string
		value        uint32
	}{
		{"3-8-3", "2-068-1", 4641}, {"8-8-8", "1-2-3", 66051},
		{"32", "4294967295", 4294967295}, {"1", "1", 1},
		{"8-8-8", "256-0-0", 16777216}, {"14", "", 16384},
		{"0", "0", 0}, {"8-8-8", "1--3", 0}, {"8-8-8", "1-٢-3", 0},
	} {
		f.Add(seed.layout, seed.text, seed.value)
	}
	f.Fuzz(func(t *testing.T, specification, input string, value uint32) {
		layout, err := pointcode.ParseLayout(specification)
		if err != nil {
			checkFuzzError(t, err)
			return
		}
		text, err := layout.Format(value)
		fits := uint64(value) < uint64(1)<<layout.Bits()
		if !fits {
			if !errors.Is(err, pointcode.ErrRange) || text != "" {
				t.Fatalf("Format(%d) failed to reject overflow: %q, %v", value, text, err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := layout.Parse(text)
			if err != nil || parsed != value {
				t.Fatalf("Format/Parse lost %d: %q, %d, %v", value, text, parsed, err)
			}
		}
		parsed, err := layout.Parse(input)
		if err != nil {
			checkFuzzError(t, err)
			if parsed != 0 {
				t.Fatal("failed parse returned nonzero value")
			}
			return
		}
		canonical, err := layout.Format(parsed)
		if err != nil {
			t.Fatal(err)
		}
		again, err := layout.Parse(canonical)
		if err != nil || again != parsed {
			t.Fatalf("Parse/Format lost %d: %q, %d, %v", parsed, canonical, again, err)
		}
		stable, err := layout.Format(again)
		if err != nil || stable != canonical {
			t.Fatalf("unstable canonical form %q -> %q, %v", canonical, stable, err)
		}
	})
}

func FuzzParseNetworkIndicator(f *testing.F) {
	for _, s := range []string{"international", "international-spare", "national", "national-reserved", "", "National", "00", "2", "national ", "international-spare\x00", "invalid NetworkIndicator(4)"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		ni, err := pointcode.ParseNetworkIndicator(s)
		if err != nil {
			if !errors.Is(err, pointcode.ErrSyntax) || ni != 0 {
				t.Fatalf("failed parse = %d, %v", ni, err)
			}
			return
		}
		if !ni.Valid() || ni.String() != s {
			t.Fatalf("accepted noncanonical text %q as %d (%q)", s, ni, ni.String())
		}
		text, err := ni.MarshalText()
		if err != nil || string(text) != s {
			t.Fatalf("marshal = %q, %v", text, err)
		}
		var decoded pointcode.NetworkIndicator
		if err := decoded.UnmarshalText(text); err != nil || decoded != ni {
			t.Fatalf("unmarshal = %d, %v", decoded, err)
		}
	})
}
