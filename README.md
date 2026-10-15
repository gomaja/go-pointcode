# go-pointcode

Go values for SS7 signalling point codes: international signalling point codes
(ISPCs), signalling area/network codes (SANCs), generic national presentation
layouts, and the network indicator. Standard library only; requires Go 1.26 or
later.

The international format implements ITU-T **Q.704 (07/96)** §2.2.3 and
**Q.708 (03/99)** §§3.4, 5.1. The package also implements the network indicator
of Q.704 §14.2.2. ITU lists both editions as in force, checked
2026-10-15. See [STANDARDS.md](STANDARDS.md) for sources, scope and corrections,
including the Q.704 Implementers' Guide.

## Install

```sh
go get github.com/gomaja/go-pointcode@main
```

`main` is the only supported version. There are no tags.

## International codes

```go
import (
    "fmt"

    "github.com/gomaja/go-pointcode"
)

p, err := pointcode.ParseISPC("2-68-1")
if err != nil {
    return err
}
fmt.Println(uint32(p), p) // 4641 2-068-1

s := p.SANC()
fmt.Println(uint16(s), s) // 580 2-068
next, err := s.ISPC(2)
if err != nil {
    return err
}
fmt.Println(next) // 2-068-2
```

An ISPC is one 14-bit pure binary value. Q.708 §5.1 represents it as NML
(0..7), K-D (000..255), and CBA (0..7), such as `2-068-1`. The first two
fields form the SANC; the third is the signalling point identification.
`String` and text marshaling pad the middle field to three digits.
Parsing also accepts one or two middle digits, with no spaces or signs.

The fields have **administrative use only**. Protocol and signalling functions
must use the whole ISPC, never its fields individually (Q.708 §5.2).
International numbering applies with network indicator `00` (Q.708 §3.4;
Q.704 §14.2.2). `NetworkIndicator` represents that context as `NIInternational`.

Q.708 §7.4 reserves SANC ranges beginning with 0 and 1; §§13.4–13.5 allow
their release. Format validation therefore accepts both ranges and zero.
Section 13.5 cautions about an ISPC with all fields zero. The package checks
format and range; it does not establish assignment or permission to use a code.

## National layouts

```go
layout, err := pointcode.ParseLayout("8-8-8")
if err != nil {
    return err
}
text, err := layout.Format(66051)
if err != nil {
    return err
}
fmt.Println(text) // 1-2-3
raw, err := layout.Parse("001-2-03")
if err != nil {
    return err
}
fmt.Println(raw) // 66051
```

Widths run from most significant to least significant, each 1..32 bits and at
most 32 bits in total. One field, such as `14`, gives plain decimal notation.
Fields allow leading zeros when parsed; formatting adds no zero padding.
Layouts are immutable values, safe to copy and share.

Q.704 §2.2.1 NOTE and §2.2.6 allow national labels with other sub-field sizes;
how such codes are written is a national convention, not an ITU-T one.
Common examples are `8-8-8` for 24-bit codes, and `7-4-5` or `4-3-4-3` for
16- and 14-bit codes, without attribution to any standard.
A `3-8-3` layout formats 4641 as `2-68-1`; use `ISPC` for canonical `2-068-1`.
There is no routing-label or service information octet codec.

## Network indicator

```go
ni, err := pointcode.ParseNetworkIndicator("national")
if err != nil {
    return err
}
fmt.Println(uint8(ni), ni) // 2 national
```

Q.704 (07/96) §14.2.2 allocates bits D-C of the sub-service field:

| Bits D C | Value | Constant | Text | Q.704 §14.2.2 meaning |
| --- | --- | --- | --- | --- |
| 0 0 | 0 | `NIInternational` | `international` | International network |
| 0 1 | 1 | `NIInternationalSpare` | `international-spare` | Spare (for international use only) |
| 1 0 | 2 | `NINational` | `national` | National network |
| 1 1 | 3 | `NINationalReserved` | `national-reserved` | Reserved for national use |

The network indicator distinguishes the numbering scheme and possibly the
label structure (Q.704 §2.4.3). ISPC numbering applies with `NIInternational`
(Q.708 §3.4). Q.704 §14.2.2 also allows the indicator to distinguish two
national signalling networks with different routing label structures; with a
single national network, the code reserved for national use can, for example,
define 16 more User Parts.
The text names are package policy; parsing accepts only the four lowercase names.

## Validation and errors

Every package error wraps `ErrSyntax` or `ErrRange` and includes operation
context. Check with `errors.Is(err, pointcode.ErrSyntax)` or
`errors.Is(err, pointcode.ErrRange)`. Parsers check complete syntax before
numeric ranges and return a zero value on error. Values are never silently masked.

Direct numeric conversions can create invalid `ISPC`, `SANC` or
`NetworkIndicator` values. Check `Valid()` before using them. Invalid `String`
output is `invalid ISPC(n)`, `invalid SANC(n)` or `invalid NetworkIndicator(n)`
with the original decimal value; `MarshalText` returns
`ErrRange` and no text. Failed `UnmarshalText` leaves the receiver unchanged;
a nil receiver returns `ErrSyntax`.

The zero `Layout` is invalid: `Bits()` is 0, `String()` is empty, and `Parse`
and `Format` return `ErrRange`.

For a `uint32` OPC/DPC from go-m3ua, check the range **before** conversion:

```go
if raw > uint32(pointcode.MaxISPC) {
    return fmt.Errorf("OPC/DPC %d: %w", raw, pointcode.ErrRange)
}
p := pointcode.ISPC(raw)
```

Converting first could truncate values above 65535 before validation.
For national numbering, keep the `uint32` and use the chosen layout.

For a `uint8` network indicator from go-m3ua protocol data, convert and check
`Valid()` because a peer can send any octet:

```go
ni := pointcode.NetworkIndicator(raw)
if !ni.Valid() {
    return fmt.Errorf("network indicator %d: %w", raw, pointcode.ErrRange)
}
```

## Migrating from go-m3ua/pc

This is a new API with explicit errors. There are no compatibility aliases.

| Old call | Replacement |
| --- | --- |
| `pc.NewPointCode(raw, pc.Variant383)` | Check `raw <= uint32(pointcode.MaxISPC)`, then `pointcode.ISPC(raw)` for international codes |
| `pc.NewPointCodeFrom(text, pc.Variant383)` | `pointcode.ParseISPC(text)` for international codes |
| `pc.Variant("8-8-8")` | `pointcode.ParseLayout("8-8-8")`; handle its error |
| `variant.BitLength()` | `layout.Bits()` after successful layout parsing |
| `pc.NewPointCode(raw, variant)` | `layout.Format(raw)`; handle overflow instead of masking it |
| `pc.NewPointCodeFrom(text, variant)` | `layout.Parse(text)`; handle syntax and range errors |
| `p.Uint32()` | `uint32(ispc)`, or the `uint32` returned by `layout.Parse` |
| `p.ConvertTo(variant)` | `otherLayout.Format(raw)`; keep the returned text explicitly |
| `p.String()` | `ispc.String()` or `layout.Format(raw)` |
| `p.Variant()` / `variant.String()` | `layout.String()` for a generic layout |

Old constructors returned nil on failure, and formatting could truncate values
or mutate a stored presentation. New constructors and parsers return errors;
formatting returns text without changing the code or layout. Layout `3-8-3`
keeps unpadded presentation; international `ISPC.String` uses Q.708 padding.

## Converter

```sh
go run ./examples/pc-conv -str 2-068-1
go run ./examples/pc-conv -raw 0x1221
go run ./examples/pc-conv -layout 8-8-8 -raw 66051
```

Supply exactly one of `-raw` or `-str`. ISPC is the default; `-layout` selects
generic fields. Raw input accepts decimal or hexadecimal with a `0x` prefix.
Zero is valid input. Exit status is 0 for success/help, 2 for invalid input,
and 1 if writing the result fails.

## License

[MIT](LICENSE), retaining the go-m3ua authors' copyright and adding the
go-pointcode authors' copyright.
