package pointcode_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gomaja/go-pointcode"
)

func TestLayoutVectors(t *testing.T) {
	for _, tt := range []struct {
		layout string
		bits   int
		raw    uint32
		text   string
	}{
		{"3-8-3", 14, 4641, "2-68-1"},
		{"3-8-3", 14, 0, "0-0-0"},
		{"3-8-3", 14, 16383, "7-255-7"},
		{"8-8-8", 24, 66051, "1-2-3"},
		{"8-8-8", 24, 16777215, "255-255-255"},
		{"7-4-5", 16, 6244, "12-3-4"},
		{"4-3-4-3", 14, 2353, "2-2-6-1"},
		{"14", 14, 4641, "4641"},
		{"1", 1, 1, "1"},
		{"31-1", 32, 4294967295, "2147483647-1"},
		{"1-31", 32, 4294967295, "1-2147483647"},
		{"32", 32, 4294967295, "4294967295"},
		{"32", 32, 0, "0"},
	} {
		t.Run(tt.layout+"/"+tt.text, func(t *testing.T) {
			layout, err := pointcode.ParseLayout(tt.layout)
			if err != nil || layout.Bits() != tt.bits || layout.String() != tt.layout {
				t.Fatalf("ParseLayout(%q) = %q, %d bits, %v", tt.layout, layout.String(), layout.Bits(), err)
			}
			text, err := layout.Format(tt.raw)
			if err != nil || text != tt.text {
				t.Fatalf("Format(%d) = %q, %v; want %q", tt.raw, text, err, tt.text)
			}
			raw, err := layout.Parse(tt.text)
			if err != nil || raw != tt.raw {
				t.Fatalf("Parse(%q) = %d, %v; want %d", tt.text, raw, err, tt.raw)
			}
		})
	}
}

func TestLayoutAllFieldWidths(t *testing.T) {
	for width := 1; width <= 32; width++ {
		layout, err := pointcode.ParseLayout(strconv.Itoa(width))
		if err != nil || layout.Bits() != width {
			t.Fatalf("width %d: %v", width, err)
		}
		maxValue := uint32((uint64(1) << width) - 1)
		want := strconv.FormatUint(uint64(maxValue), 10)
		text, err := layout.Format(maxValue)
		if err != nil || text != want {
			t.Fatalf("width %d max: %q, %v", width, text, err)
		}
		raw, err := layout.Parse("000" + want)
		if err != nil || raw != maxValue {
			t.Fatalf("width %d parse max = %d, %v", width, raw, err)
		}
		_, err = layout.Parse(strconv.FormatUint(uint64(maxValue)+1, 10))
		requireError(t, err, pointcode.ErrRange, "parse")
		if width < 32 {
			text, err := layout.Format(maxValue + 1)
			requireError(t, err, pointcode.ErrRange, "format")
			if text != "" {
				t.Fatalf("overflow formatted %q", text)
			}
		}
	}
}

func TestLayoutManyFieldsAndLeadingZeros(t *testing.T) {
	widths := strings.TrimSuffix(strings.Repeat("1-", 32), "-")
	want := strings.TrimSuffix(strings.Repeat("1-", 32), "-")
	layout, err := pointcode.ParseLayout(widths)
	if err != nil || layout.Bits() != 32 || layout.String() != widths {
		t.Fatalf("32 fields: %q, %d, %v", layout.String(), layout.Bits(), err)
	}
	text, err := layout.Format(4294967295)
	if err != nil || text != want {
		t.Fatalf("32 fields max = %q, %v", text, err)
	}
	raw, err := layout.Parse(want)
	if err != nil || raw != 4294967295 {
		t.Fatalf("32 fields parse = %d, %v", raw, err)
	}
	for _, specification := range []string{"08-008-8", strings.Repeat("0", 8192) + "8-8-8"} {
		layout, err := pointcode.ParseLayout(specification)
		if err != nil || layout.String() != "8-8-8" {
			t.Fatalf("leading zeros in layout = %q, %v", layout.String(), err)
		}
		input := strings.Repeat("0", 8192) + "1-002-03"
		raw, err := layout.Parse(input)
		if err != nil || raw != 66051 {
			t.Fatalf("leading zeros in fields = %d, %v", raw, err)
		}
		canonical, err := layout.Format(raw)
		if err != nil || canonical != "1-2-3" {
			t.Fatalf("leading zero canonicalization = %q, %v", canonical, err)
		}
	}
}

func TestParseLayoutRejections(t *testing.T) {
	for _, specification := range []string{"", "-", "8-", "-8", "8--8", "+8", " 8", "8 ", "8\n", "a", "８", "8-٨", "0x8", "8.8.8", "8-0_8", "33-x"} {
		t.Run("syntax/"+specification, func(t *testing.T) {
			layout, err := pointcode.ParseLayout(specification)
			requireError(t, err, pointcode.ErrSyntax, "parse layout")
			if layout.Bits() != 0 || layout.String() != "" {
				t.Fatalf("failed parse produced a layout %q", layout.String())
			}
		})
	}
	for _, specification := range []string{"0", "00", "33", "999999999999999999999999", "16-17", "32-1", "8-0-8", strings.TrimSuffix(strings.Repeat("1-", 33), "-")} {
		layout, err := pointcode.ParseLayout(specification)
		requireError(t, err, pointcode.ErrRange, "parse layout")
		if layout.Bits() != 0 || layout.String() != "" {
			t.Fatalf("failed parse produced a layout %q", layout.String())
		}
	}
}

func TestLayoutParseRejections(t *testing.T) {
	layout, err := pointcode.ParseLayout("8-8-8")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"", "1", "1-2", "1-2-3-4", "-2-3", "1--3", "1-2-", "-1-2-3", "+1-2-3", "1-+2-3", "1-2-+3",
		" 1-2-3", "1- 2-3", "1-2-3 ", "1-2-3\n", "a-2-3", "1-a-3", "1-2-a", "１-2-3", "1-٢-3", "1-2-３",
		"0x1-2-3", "1-0_2-3", "1-\x002-3", "256-x-0",
	} {
		t.Run("syntax/"+text, func(t *testing.T) {
			raw, err := layout.Parse(text)
			requireError(t, err, pointcode.ErrSyntax, "parse")
			if raw != 0 {
				t.Fatalf("failed parse returned %d", raw)
			}
		})
	}
	for _, text := range []string{"256-0-0", "0-256-0", "0-0-256", "4294967296-0-0", "999999999999999999999999-0-0"} {
		raw, err := layout.Parse(text)
		requireError(t, err, pointcode.ErrRange, "parse")
		if raw != 0 {
			t.Fatalf("failed parse returned %d", raw)
		}
	}
}

func TestZeroLayout(t *testing.T) {
	var layout pointcode.Layout
	if layout.Bits() != 0 || layout.String() != "" {
		t.Fatalf("zero layout: %d, %q", layout.Bits(), layout.String())
	}
	for _, raw := range []uint32{0, 1, 4294967295} {
		text, err := layout.Format(raw)
		requireError(t, err, pointcode.ErrRange, "format")
		if text != "" {
			t.Fatalf("zero layout formatted %q", text)
		}
	}
	for _, text := range []string{"", "0", "1-2-3"} {
		raw, err := layout.Parse(text)
		requireError(t, err, pointcode.ErrRange, "parse")
		if raw != 0 {
			t.Fatalf("zero layout parsed %d", raw)
		}
	}
}
