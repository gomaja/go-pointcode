package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name   string
		args   []string
		status int
		output string
		err    string
	}{
		{"parse ISPC", []string{"-str", "2-68-1"}, 0, "Raw: 4641, Formatted: 2-068-1, Layout: ISPC\n", ""},
		{"raw ISPC", []string{"-raw", "4641"}, 0, "Raw: 4641, Formatted: 2-068-1, Layout: ISPC\n", ""},
		{"hex ISPC", []string{"-raw", "0x1221"}, 0, "Raw: 4641, Formatted: 2-068-1, Layout: ISPC\n", ""},
		{"raw zero", []string{"-raw", "0"}, 0, "Raw: 0, Formatted: 0-000-0, Layout: ISPC\n", ""},
		{"zero ISPC", []string{"-str", "0-000-0"}, 0, "Raw: 0, Formatted: 0-000-0, Layout: ISPC\n", ""},
		{"max ISPC", []string{"-raw", "16383"}, 0, "Raw: 16383, Formatted: 7-255-7, Layout: ISPC\n", ""},
		{"national raw", []string{"-raw", "66051", "-layout", "8-8-8"}, 0, "Raw: 66051, Formatted: 1-2-3, Layout: 8-8-8\n", ""},
		{"national string", []string{"-str", "001-2-03", "-layout", "8-8-8"}, 0, "Raw: 66051, Formatted: 1-2-3, Layout: 8-8-8\n", ""},
		{"uint32 max", []string{"-raw", "4294967295", "-layout", "32"}, 0, "Raw: 4294967295, Formatted: 4294967295, Layout: 32\n", ""},
		{"decimal leading zero", []string{"-raw", "04641"}, 0, "Raw: 4641, Formatted: 2-068-1, Layout: ISPC\n", ""},
		{"layout canonical", []string{"-str", "4641", "-layout", "014"}, 0, "Raw: 4641, Formatted: 4641, Layout: 14\n", ""},
		{"help", []string{"-h"}, 0, "", "Usage of pc-conv:"},
		{"no input", nil, 2, "", "exactly one of -raw or -str"},
		{"both inputs", []string{"-raw", "0", "-str", "0-000-0"}, 2, "", "exactly one of -raw or -str"},
		{"empty raw", []string{"-raw", ""}, 2, "", "invalid raw"},
		{"empty string", []string{"-str", ""}, 2, "", "parse ISPC"},
		{"repeated raw", []string{"-raw", "0", "-raw", "1"}, 2, "", "repeated flag -raw"},
		{"repeated str", []string{"-str", "2-068-1", "-str", "2-068-2"}, 2, "", "repeated flag -str"},
		{"repeated layout", []string{"-raw", "0", "-layout", "14", "-layout", "32"}, 2, "", "repeated flag -layout"},
		{"unknown flag", []string{"-bad"}, 2, "", "flag provided but not defined"},
		{"missing flag argument", []string{"-raw"}, 2, "", "flag needs an argument"},
		{"positional argument", []string{"-raw", "0", "extra"}, 2, "", "unexpected positional argument"},
		{"raw negative", []string{"-raw", "-1"}, 2, "", "invalid raw"},
		{"raw overflow", []string{"-raw", "4294967296", "-layout", "32"}, 2, "", "invalid raw"},
		{"raw junk", []string{"-raw", "x"}, 2, "", "invalid raw"},
		{"raw binary", []string{"-raw", "0b1"}, 2, "", "invalid raw"},
		{"bad ISPC syntax", []string{"-str", "2--1"}, 2, "", "parse ISPC"},
		{"bad ISPC range", []string{"-str", "2-256-1"}, 2, "", "parse ISPC"},
		{"ISPC raw overflow", []string{"-raw", "16384"}, 2, "", "ISPC range"},
		{"ISPC truncation", []string{"-raw", "65536"}, 2, "", "ISPC range"},
		{"invalid layout", []string{"-raw", "0", "-layout", "0"}, 2, "", "parse layout"},
		{"bad national syntax", []string{"-str", "1-2", "-layout", "8-8-8"}, 2, "", "parse"},
		{"bad national range", []string{"-str", "256-2-3", "-layout", "8-8-8"}, 2, "", "parse"},
		{"national raw overflow", []string{"-raw", "16777216", "-layout", "8-8-8"}, 2, "", "format"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(tt.args, &stdout, &stderr)
			if status != tt.status || stdout.String() != tt.output {
				t.Fatalf("status/output = %d/%q, want %d/%q; stderr=%q", status, stdout.String(), tt.status, tt.output, stderr.String())
			}
			if tt.err == "" {
				if stderr.Len() != 0 {
					t.Fatalf("unexpected stderr: %q", stderr.String())
				}
			} else if !strings.Contains(stderr.String(), tt.err) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.err)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	if status := run([]string{"-raw", "0"}, failingWriter{}, &stderr); status != 1 || !strings.Contains(stderr.String(), "write failed") {
		t.Fatalf("output failure: status=%d, stderr=%q", status, stderr.String())
	}
}
