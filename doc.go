// Package pointcode represents SS7 signalling point codes (SPCs), which identify
// signalling points in a signalling network (Q.708 (03/99) §3.3), and the
// network indicator, which tells which numbering a code belongs to
// (Q.704 (07/96) §2.4.3).
//
// ISPC and SANC implement the international code format in ITU-T Q.704 (07/96)
// §2.2.3 and Q.708 (03/99) §§3.4, 5.1. NetworkIndicator implements the network
// indicator of Q.704 §14.2.2. International numbering, and with it ISPC,
// applies to messages whose network indicator is NIInternational, coded 00
// (Q.708 §3.4; Q.704 §2.4.3). National networks number and present their codes
// in their own way; Layout provides generic presentation fields under Q.704
// §2.2.1 NOTE and §2.2.6, without implementing a national standard or
// routing-label codec.
//
// The ISPC fields have administrative use only and must not be used individually
// by protocol or signalling functions (Q.708 §5.2). Ranges 0 and 1 are reserved
// but may be released (Q.708 §§7.4, 13.4–13.5); format validation accepts them,
// including zero. Assignment and authorization to use a code are separate.
//
// To convert a uint32 OPC or DPC from go-m3ua to ISPC, first check that it is no
// greater than uint32(MaxISPC), then convert with ISPC(raw). Checking after a
// conversion to the narrower uint16 type may miss truncation. For a national
// code, choose its presentation layout and pass the uint32 to Layout.Format.
// A uint8 network indicator from go-m3ua protocol data converts with
// NetworkIndicator(raw); check Valid afterwards, because a peer can send any
// octet.
package pointcode
