package pointcode_test

import (
	"fmt"

	"github.com/gomaja/go-pointcode"
)

func ExampleParseISPC() {
	p, err := pointcode.ParseISPC("2-68-1")
	if err != nil {
		panic(err)
	}
	fmt.Println(uint16(p), p)
	// Output: 4641 2-068-1
}

func ExampleISPC_SANC() {
	p, err := pointcode.NewISPC(2, 68, 1)
	if err != nil {
		panic(err)
	}
	s := p.SANC()
	next, err := s.ISPC(2)
	if err != nil {
		panic(err)
	}
	fmt.Println(uint16(s), s, next)
	// Output: 580 2-068 2-068-2
}

func ExampleLayout_Format() {
	layout, err := pointcode.ParseLayout("8-8-8")
	if err != nil {
		panic(err)
	}
	text, err := layout.Format(66051)
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
	// Output: 1-2-3
}
