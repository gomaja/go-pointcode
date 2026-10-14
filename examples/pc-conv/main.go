// pc-conv prints a raw and formatted signalling point code. By default it uses
// Q.708 ISPC notation; -layout selects generic national presentation fields.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gomaja/go-pointcode"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("pc-conv", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var rawText, codeText, layoutText string
	var rawSet, codeSet, layoutSet bool
	register := func(name, usage string, target *string, seen *bool) {
		flags.Func(name, usage, func(value string) error {
			if *seen {
				return fmt.Errorf("repeated flag -%s", name)
			}
			*seen, *target = true, value
			return nil
		})
	}
	register("raw", "Raw point code in decimal or hexadecimal (0x prefix).", &rawText, &rawSet)
	register("str", "Formatted point code, with decimal fields separated by hyphens.", &codeText, &codeSet)
	register("layout", "National field widths, most significant first (for example 8-8-8).", &layoutText, &layoutSet)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	report := func(err error, status int) int {
		// Reporting is best effort if the diagnostic stream itself has failed.
		_, _ = fmt.Fprintln(stderr, err)
		return status
	}
	if flags.NArg() != 0 {
		return report(errors.New("unexpected positional argument"), 2)
	}
	if rawSet == codeSet {
		return report(errors.New("exactly one of -raw or -str is required"), 2)
	}
	var raw uint32
	if rawSet {
		base, digits := 10, rawText
		if strings.HasPrefix(rawText, "0x") || strings.HasPrefix(rawText, "0X") {
			base, digits = 16, rawText[2:]
		}
		value, err := strconv.ParseUint(digits, base, 32)
		if err != nil {
			return report(fmt.Errorf("invalid raw point code %q: %w", rawText, err), 2)
		}
		raw = uint32(value)
	}
	var formatted, name string
	if layoutSet {
		layout, err := pointcode.ParseLayout(layoutText)
		if err != nil {
			return report(err, 2)
		}
		if codeSet {
			raw, err = layout.Parse(codeText)
			if err != nil {
				return report(err, 2)
			}
		}
		formatted, err = layout.Format(raw)
		if err != nil {
			return report(err, 2)
		}
		name = layout.String()
	} else {
		var code pointcode.ISPC
		if codeSet {
			var err error
			code, err = pointcode.ParseISPC(codeText)
			if err != nil {
				return report(err, 2)
			}
			raw = uint32(code)
		} else {
			// Check before the uint16 conversion; Q.704 (07/96) §2.2.3,
			// Figure 3 and Q.708 (03/99) §3.4 specify a 14-bit ISPC.
			if raw > uint32(pointcode.MaxISPC) {
				return report(fmt.Errorf("raw %d exceeds ISPC range: %w", raw, pointcode.ErrRange), 2)
			}
			code = pointcode.ISPC(raw)
		}
		formatted, name = code.String(), "ISPC"
	}
	if _, err := fmt.Fprintf(stdout, "Raw: %d, Formatted: %s, Layout: %s\n", raw, formatted, name); err != nil {
		return report(err, 1)
	}
	return 0
}
