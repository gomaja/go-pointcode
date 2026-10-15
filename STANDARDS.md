# Standards and scope

Publication status checked against ITU's recommendation indexes and edition
pages on **2026-10-15**. These are ITU-T Recommendations; RFC Editor and IETF
Datatracker relationship checks do not establish their status.

| Document | Edition | ITU status | Authority |
| --- | --- | --- | --- |
| Q.704, *Signalling network functions and messages* | 07/96 | In force | [Index](https://www.itu.int/rec/T-REC-Q.704/en), [edition](https://www.itu.int/rec/T-REC-Q.704-199607-I/en) |
| Q.708, *Assignment procedures for international signalling point codes* | 03/99 | In force | [Index](https://www.itu.int/rec/T-REC-Q.708/en), [edition](https://www.itu.int/rec/T-REC-Q.708-199903-I/en) |

Both indexes list their 1988 and 1993 editions as superseded. The implementation
uses the in-force editions above. Recheck the indexes, edition pages and
corrections before a significant revisit of this subsystem.

## Corrections

- **Q.704:** the
  [Implementers' Guide for Q.704 (07/96)](https://www.itu.int/rec/T-REC-Q.Imp704/en)
  has three editions, 03/99, 12/99 and 12/00, all listed as in force. The 12/00
  edition states that it incorporates the 12/99 edition, which incorporated the
  03/99 one. It allocates further service indicators, adds a service indicator
  extension (service indicator 1111 extends the routing label by two octets,
  leaving room for a later point-code extension that it does not define),
  clarifies link set emergency restart, rewords clause 15.2 and corrects SDL
  figures. It replaces §14.2.1, the service indicator. It does not change the
  network indicator of §14.2.2, the 14-bit point code of §2.2.3 or anything
  else this package implements.
- **Q.704:** the ITU index also lists a covering note dated **17.09.99**,
  described as an English-only erratum. ITU offers no document for it: the
  index, the 07/96 edition page, the newer Q.704 recommendation record, the
  Study Group 11 report archive and ITU's search were checked on 2026-10-14
  without finding one. Its content could therefore not be reviewed. The 12/00
  Implementers' Guide, which postdates it, reports corrections of editorial
  errors in the SDL and text.
- **Q.708:** no correction or erratum is listed on its recommendation index
  or edition page as checked on 2026-10-14. This is a statement about those
  pages, not a claim that no other published clarification exists.

No complete implementation of either Recommendation is claimed. The package
implements the representation and presentation of point codes and of the
network indicator. It does not implement assignment administration, MTP3
signalling functions, or a routing-label or service information octet codec.

## Clause-to-code map

| Source in the edition above | Requirement or boundary | Implementation and evidence |
| --- | --- | --- |
| Q.704 §2.2.3, Figure 3 | DPC/OPC pure binary; 14-bit international code fields | `ISPC`, `MaxISPC`, checked numeric conversion in `doc.go` and `examples/pc-conv`; literal vectors and exhaustive `TestAllISPCs` |
| Q.704 §2.2.1 NOTE; §2.2.6 | Modified national labels may have other sub-field sizes | `Layout`, `ParseLayout`, `Format`, `Parse`; `TestLayoutVectors`, `TestLayoutAllFieldWidths` |
| Q.704 §14.2.2 | Four codes, bits D-C of the sub-service field | `NetworkIndicator`, its constants, `ParseNetworkIndicator`, `Valid`, `String`, text marshaling; `TestNetworkIndicatorCodes`, `TestAllNetworkIndicators`, `FuzzParseNetworkIndicator` |
| Q.704 §2.4.3; Q.708 §§1, 3.4 | Network indicator selects the numbering; ISPC with NI=00 | `NetworkIndicator` and `ISPC` documentation, README |
| Q.708 §3.3; §3.4 | SPC identifies a signalling point; ISPC is a unique 14-bit format used with NI=00 | Package documentation and `ISPC`; `TestISPCVectors`, `TestInvalidValues` |
| Q.708 §5.1, Figure 1 | NML: 3 bits; K-D: 8 bits; CBA: 3 bits; three decimal numbers, e.g. 2-068-1 | `NewISPC`, `ParseISPC`, field accessors, `String`, text marshaling; `TestAllISPCs` |
| Q.708 §5.1, Figure 1 | NML and K-D together form the SANC | `SANC`, `MaxSANC`, `NewSANC`, `ParseSANC`, `ISPC.SANC`; `TestSANCVectors`, `TestAllSANCs` |
| Q.708 §5.2 | Fields have administrative use only, never individual protocol use | Type and package documentation, README; accessors support administration/display |
| Q.708 §7.2 | A SANC with eight signalling point identifications gives eight ISPCs | `SANC.ISPC`; all eight combinations for every SANC in `TestAllSANCs` |
| Q.708 §7.4; §13.4–§13.5 | Ranges 0 and 1 reserved with possible release; caution with all-zero ISPC | Type/package documentation; both ranges and zero accepted in exhaustive tests |

## Text and range policy

`NetworkIndicator` uses the names `international`, `international-spare`,
`national` and `national-reserved`. Exact lowercase parsing with no numeric
text, invalid-value formatting and atomic decoding are package policy;
Q.704 §14.2.2 defines the codes and their meanings only.

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
