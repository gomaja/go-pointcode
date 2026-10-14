# Standards and scope

Publication status checked against ITU's recommendation indexes and edition
pages on **2026-10-14**. These are ITU-T Recommendations; RFC Editor and IETF
Datatracker relationship checks do not establish their status.

| Document | Edition | ITU status | Authority |
| --- | --- | --- | --- |
| Q.704, *Signalling network functions and messages* | 07/96 | In force | [Index](https://www.itu.int/rec/T-REC-Q.704/en), [edition](https://www.itu.int/rec/T-REC-Q.704-199607-I/en) |
| Q.708, *Assignment procedures for international signalling point codes* | 03/99 | In force | [Index](https://www.itu.int/rec/T-REC-Q.708/en), [edition](https://www.itu.int/rec/T-REC-Q.708-199903-I/en) |

Both indexes list their 1988 and 1993 editions as superseded. The implementation
uses the in-force editions above. Recheck the indexes, edition pages and
corrections before a significant revisit of this subsystem.

## Corrections

- **Q.704:** the ITU index lists a covering note dated **17.09.99**, described
  as an English-only erratum. **ERRATUM-PENDING** — its text and applicability
  remain under review; the correction's status will be filled in separately.
- **Q.708:** no correction or erratum is listed on its recommendation index
  or edition page as checked on 2026-10-14. This is a statement about those
  pages, not a claim that no other published clarification exists.

No complete implementation of either Recommendation is claimed. The package
implements code representation and presentation, with the correction review
above still pending. It does not implement assignment administration, MTP3
signalling functions, a routing-label codec, or a network-indicator type.

## Clause-to-code map

| Source in the edition above | Requirement or boundary | Implementation and evidence |
| --- | --- | --- |
| Q.704 §2.2.3, Figure 3 | DPC/OPC pure binary; 14-bit international code fields | `ISPC`, `MaxISPC`, checked numeric conversion in `doc.go` and `examples/pc-conv`; literal vectors and exhaustive `TestAllISPCs` |
| Q.704 §2.2.1 NOTE; §2.2.6 | Modified national labels may have other sub-field sizes | `Layout`, `ParseLayout`, `Format`, `Parse`; `TestLayoutVectors`, `TestLayoutAllFieldWidths` |
| Q.704 §2.4.3; §14.2.2 | Network indicator distinguishes numbering schemes; 00 is international | `doc.go`, README; the caller supplies network context |
| Q.708 §3.3; §3.4 | SPC identifies a signalling point; ISPC is a unique 14-bit format used with NI=00 | Package documentation and `ISPC`; `TestISPCVectors`, `TestInvalidValues` |
| Q.708 §5.1, Figure 1 | NML: 3 bits; K-D: 8 bits; CBA: 3 bits; three decimal numbers, e.g. 2-068-1 | `NewISPC`, `ParseISPC`, field accessors, `String`, text marshaling; `TestAllISPCs` |
| Q.708 §5.1, Figure 1 | NML and K-D together form the SANC | `SANC`, `MaxSANC`, `NewSANC`, `ParseSANC`, `ISPC.SANC`; `TestSANCVectors`, `TestAllSANCs` |
| Q.708 §5.2 | Fields have administrative use only, never individual protocol use | Type and package documentation, README; accessors support administration/display |
| Q.708 §7.2 | A SANC with eight signalling point identifications gives eight ISPCs | `SANC.ISPC`; all eight combinations for every SANC in `TestAllSANCs` |
| Q.708 §7.4; §13.4–§13.5 | Ranges 0 and 1 reserved with possible release; caution with all-zero ISPC | Type/package documentation; both ranges and zero accepted in exhaustive tests |

## Text and range policy

`ISPC.String` uses the three-digit middle field shown by Q.708 §5.1.
`ParseISPC` additionally accepts one or two middle digits; `ParseSANC` applies
the same rule to the first two fields. Parsers require ASCII decimal digits,
one digit for NML/CBA, no spaces/signs, and complete syntax before range checks.
These strict acceptance rules and the error API are package policy.

National presentation is a national convention, not an ITU-T one. The generic
layouts describe widths only and do not claim conformance to a national
standard. Common examples are 8-8-8 (24 bits), 7-4-5 (16 bits), and 4-3-4-3
(14 bits), without attributing them to any standard. Generic 3-8-3 presentation
is unpadded (`2-68-1`), whereas the ISPC canonical form is `2-068-1`.

The generic 1..32-bit limit, leading-zero handling, immutable layout storage,
sentinel errors, invalid-value formatting, and atomic text decoding are
package API choices rather than claims about ITU-T national presentation.
No allocation database or national numbering plan is included.
