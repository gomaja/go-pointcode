package pointcode_test

import (
	"encoding"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gomaja/go-pointcode"
)

var (
	_ encoding.TextMarshaler   = pointcode.ISPC(0)
	_ encoding.TextUnmarshaler = (*pointcode.ISPC)(nil)
	_ encoding.TextMarshaler   = pointcode.SANC(0)
	_ encoding.TextUnmarshaler = (*pointcode.SANC)(nil)
)

func requireError(t *testing.T, err, class error, context string) {
	t.Helper()
	if !errors.Is(err, class) || err == class {
		t.Fatalf("error = %v, want wrapped %v", err, class)
	}
	if !strings.HasPrefix(err.Error(), "pointcode: ") || !strings.Contains(err.Error(), context) {
		t.Fatalf("error %q lacks operation context %q", err, context)
	}
}

// Q.708 (03/99) §5.1, Figure 1: these literal values pin the binary format.
func TestISPCVectors(t *testing.T) {
	for _, tt := range []struct {
		text      string
		canonical string
		value     pointcode.ISPC
	}{
		{"2-068-1", "2-068-1", 4641},
		{"2-68-1", "2-068-1", 4641},
		{"0-0-0", "0-000-0", 0},
		{"0-000-0", "0-000-0", 0},
		{"1-255-7", "1-255-7", 4095},
		{"7-255-7", "7-255-7", 16383},
		{"2-01-1", "2-001-1", 4105},
	} {
		t.Run(tt.text, func(t *testing.T) {
			p, err := pointcode.ParseISPC(tt.text)
			if err != nil || p != tt.value || p.String() != tt.canonical {
				t.Fatalf("ParseISPC(%q) = %d, %v, %q", tt.text, p, err, p.String())
			}
		})
	}
	p, err := pointcode.NewISPC(2, 68, 1)
	if err != nil || p != 4641 || p.Zone() != 2 || p.AreaNetwork() != 68 || p.SignallingPoint() != 1 || p.SANC() != 580 {
		t.Fatalf("NewISPC(2,68,1) = %d, %v; fields %d/%d/%d, SANC %d", p, err, p.Zone(), p.AreaNetwork(), p.SignallingPoint(), p.SANC())
	}
}

// These loops catch reversed fields, wrong masks, skipped limits and lossy text.
func TestAllISPCs(t *testing.T) {
	for raw := 0; raw <= 16383; raw++ {
		p := pointcode.ISPC(raw)
		zone, area, sp := uint8(raw/2048), uint8((raw/8)%256), uint8(raw%8)
		want := fmt.Sprintf("%d-%03d-%d", zone, area, sp)
		if !p.Valid() || p.Zone() != zone || p.AreaNetwork() != area || p.SignallingPoint() != sp || p.SANC() != pointcode.SANC(raw/8) || p.String() != want {
			t.Fatalf("ISPC %d: fields or notation disagree with Q.708 Figure 1", raw)
		}
		created, err := pointcode.NewISPC(zone, area, sp)
		if err != nil || created != p {
			t.Fatalf("NewISPC for %d = %d, %v", raw, created, err)
		}
		parsed, err := pointcode.ParseISPC(want)
		if err != nil || parsed != p {
			t.Fatalf("ParseISPC(%q) = %d, %v", want, parsed, err)
		}
		text, err := p.MarshalText()
		if err != nil || string(text) != want {
			t.Fatalf("MarshalText(%d) = %q, %v", raw, text, err)
		}
		var decoded pointcode.ISPC
		if err := decoded.UnmarshalText(text); err != nil || decoded != p {
			t.Fatalf("UnmarshalText(%q) = %d, %v", text, decoded, err)
		}
	}
}

func TestISPCRejections(t *testing.T) {
	for _, s := range []string{
		"", "2", "2-068", "2-068-1-0", "2--1", "-068-1", "2-068-", "-2-068-1", "2-068--1",
		"+2-068-1", "2-+68-1", "2-068-+1", " 2-068-1", "2- 68-1", "2-068-1 ", "2-068-1\n",
		"02-068-1", "2-068-01", "2-0068-1", "x-068-1", "2-x-1", "2-068-x",
		"２-068-1", "2-٠٦٨-1", "2-068-１", "2.068.1", "2-0x44-1", "2-0_68-1", "2-\x0068-1",
		"8-068-x", "2-999-xx",
	} {
		t.Run("syntax/"+s, func(t *testing.T) {
			p, err := pointcode.ParseISPC(s)
			requireError(t, err, pointcode.ErrSyntax, "parse ISPC")
			if p != 0 {
				t.Fatalf("invalid parse returned %d", p)
			}
		})
	}
	for _, s := range []string{"8-068-1", "9-068-1", "2-256-1", "2-999-1", "2-068-8", "2-068-9"} {
		t.Run("range/"+s, func(t *testing.T) {
			p, err := pointcode.ParseISPC(s)
			requireError(t, err, pointcode.ErrRange, "parse ISPC")
			if p != 0 {
				t.Fatalf("invalid parse returned %d", p)
			}
		})
	}
	for _, fields := range [][3]uint8{{8, 0, 0}, {255, 0, 0}, {0, 0, 8}, {0, 255, 255}} {
		p, err := pointcode.NewISPC(fields[0], fields[1], fields[2])
		requireError(t, err, pointcode.ErrRange, "new ISPC")
		if p != 0 {
			t.Fatalf("invalid constructor returned %d", p)
		}
	}
}

func TestSANCVectors(t *testing.T) {
	for _, tt := range []struct {
		text      string
		canonical string
		value     pointcode.SANC
	}{
		{"2-068", "2-068", 580}, {"2-68", "2-068", 580},
		{"0-0", "0-000", 0}, {"0-000", "0-000", 0},
		{"1-255", "1-255", 511}, {"7-255", "7-255", 2047}, {"2-01", "2-001", 513},
	} {
		t.Run(tt.text, func(t *testing.T) {
			s, err := pointcode.ParseSANC(tt.text)
			if err != nil || s != tt.value || s.String() != tt.canonical {
				t.Fatalf("ParseSANC(%q) = %d, %v, %q", tt.text, s, err, s.String())
			}
		})
	}
}

func TestAllSANCs(t *testing.T) {
	for raw := 0; raw <= 2047; raw++ {
		s := pointcode.SANC(raw)
		zone, area := uint8(raw/256), uint8(raw%256)
		want := fmt.Sprintf("%d-%03d", zone, area)
		if !s.Valid() || s.Zone() != zone || s.AreaNetwork() != area || s.String() != want {
			t.Fatalf("SANC %d: fields or notation disagree with Q.708 Figure 1", raw)
		}
		created, err := pointcode.NewSANC(zone, area)
		if err != nil || created != s {
			t.Fatalf("NewSANC for %d = %d, %v", raw, created, err)
		}
		parsed, err := pointcode.ParseSANC(want)
		if err != nil || parsed != s {
			t.Fatalf("ParseSANC(%q) = %d, %v", want, parsed, err)
		}
		text, err := s.MarshalText()
		if err != nil || string(text) != want {
			t.Fatalf("MarshalText(%d) = %q, %v", raw, text, err)
		}
		var decoded pointcode.SANC
		if err := decoded.UnmarshalText(text); err != nil || decoded != s {
			t.Fatalf("UnmarshalText(%q) = %d, %v", text, decoded, err)
		}
		// Q.708 §7.2: each SANC gives exactly eight distinct ISPCs.
		for sp := uint8(0); sp < 8; sp++ {
			p, err := s.ISPC(sp)
			if err != nil || p != pointcode.ISPC(raw*8+int(sp)) {
				t.Fatalf("SANC %d, point %d = %d, %v", raw, sp, p, err)
			}
		}
	}
}

func TestSANCRejections(t *testing.T) {
	for _, text := range []string{
		"", "2", "2-068-1", "-068", "2-", "2--068", "-2-068", "+2-068", "2-+68",
		" 2-068", "2- 68", "2-068 ", "2-068\n", "02-068", "2-0068", "x-068", "2-x",
		"２-068", "2-٠٦٨", "2.068", "2-0x44", "2-0_68", "8-x",
	} {
		t.Run("syntax/"+text, func(t *testing.T) {
			s, err := pointcode.ParseSANC(text)
			requireError(t, err, pointcode.ErrSyntax, "parse SANC")
			if s != 0 {
				t.Fatalf("invalid parse returned %d", s)
			}
		})
	}
	for _, text := range []string{"8-068", "9-068", "2-256", "2-999"} {
		s, err := pointcode.ParseSANC(text)
		requireError(t, err, pointcode.ErrRange, "parse SANC")
		if s != 0 {
			t.Fatalf("invalid parse returned %d", s)
		}
	}
	for _, zone := range []uint8{8, 255} {
		s, err := pointcode.NewSANC(zone, 255)
		requireError(t, err, pointcode.ErrRange, "new SANC")
		if s != 0 {
			t.Fatalf("invalid constructor returned %d", s)
		}
	}
	for _, sp := range []uint8{8, 255} {
		p, err := pointcode.SANC(580).ISPC(sp)
		requireError(t, err, pointcode.ErrRange, "SANC")
		if p != 0 {
			t.Fatalf("invalid point returned %d", p)
		}
	}
}

func TestInvalidValues(t *testing.T) {
	for raw := 16384; raw <= 65535; raw++ {
		p := pointcode.ISPC(raw)
		if p.Valid() || p.String() != fmt.Sprintf("invalid ISPC(%d)", raw) {
			t.Fatalf("invalid ISPC %d looks valid: %q", raw, p.String())
		}
		text, err := p.MarshalText()
		requireError(t, err, pointcode.ErrRange, "marshal ISPC")
		if text != nil {
			t.Fatalf("invalid ISPC marshaled %q", text)
		}
	}
	for raw := 2048; raw <= 65535; raw++ {
		s := pointcode.SANC(raw)
		if s.Valid() || s.String() != fmt.Sprintf("invalid SANC(%d)", raw) {
			t.Fatalf("invalid SANC %d looks valid: %q", raw, s.String())
		}
		text, err := s.MarshalText()
		requireError(t, err, pointcode.ErrRange, "marshal SANC")
		if text != nil {
			t.Fatalf("invalid SANC marshaled %q", text)
		}
		p, err := s.ISPC(0)
		requireError(t, err, pointcode.ErrRange, "SANC")
		if p != 0 {
			t.Fatalf("invalid SANC produced ISPC %d", p)
		}
	}
}

func TestTextUnmarshal(t *testing.T) {
	p, s := pointcode.ISPC(4641), pointcode.SANC(580)
	for _, tt := range []struct {
		ispc, sanc string
		class      error
	}{
		{"bad", "bad", pointcode.ErrSyntax},
		{"8-068-1", "8-068", pointcode.ErrRange},
	} {
		requireError(t, p.UnmarshalText([]byte(tt.ispc)), tt.class, "ISPC")
		requireError(t, s.UnmarshalText([]byte(tt.sanc)), tt.class, "SANC")
		if p != 4641 || s != 580 {
			t.Fatal("failed unmarshal changed receiver")
		}
	}
	if err := p.UnmarshalText([]byte("2-68-2")); err != nil || p != 4642 {
		t.Fatalf("ISPC unmarshal = %d, %v", p, err)
	}
	if err := s.UnmarshalText([]byte("2-69")); err != nil || s != 581 {
		t.Fatalf("SANC unmarshal = %d, %v", s, err)
	}
	var nilISPC *pointcode.ISPC
	var nilSANC *pointcode.SANC
	requireError(t, nilISPC.UnmarshalText([]byte("2-068-1")), pointcode.ErrSyntax, "ISPC")
	requireError(t, nilSANC.UnmarshalText([]byte("2-068")), pointcode.ErrSyntax, "SANC")
}
