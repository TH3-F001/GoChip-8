package main

import (
	// "github.com/TH3-F001/GoChip-8/pkg/display/ansi"
	"github.com/TH3-F001/GoChip-8/pkg/display/ncurses"
)

var memory [4096]byte = [4096]byte{}
var stack [16]uint16 = [16]uint16{}
var v [16]byte = [16]byte{} // Variable Registers
var pc uint16               // Program Counter
var sp uint16               // Stack Pointer
var ir uint16               // Index Register
var dt byte                 // Delay Timer
var st byte                 // Sound Timer

func main() {
	// for index, address :=what if i  range memory {
	// 	fmt.Println(index, address)
	// }
	for {
		ansi.Test()
	}
}
