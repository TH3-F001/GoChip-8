package main

import (
	"log"
	"math"
	"math/rand/v2"

	"github.com/TH3-F001/GoChip-8/chip8/pkg/io"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/dataconverter"
)

// I am aware that this abstracts my code, in some cases needlessly, but in the end the point was to be able to quickly isolate
// a problem instruction and work on it in a file which didnt contend with the screen space of the switch statement. Sue me.

var rShiftMap map[bool]func(uint16) = map[bool]func(uint16){
	true: func(opcode uint16) { // Cosmac compatible right shift function
		x := getOpcodeNibble(opcode, 1)
		y := getOpcodeNibble(opcode, 2)
		V[x] = V[y]
		V[0xF] = (V[x] & 0x1)
		V[x] = V[x] >> 1
	},
	false: func(opcode uint16) { // Chip-48/Super-Chip compatible right shift function
		x := getOpcodeNibble(opcode, 1)
		V[0xF] = (V[x] & 0x1)
		V[x] = V[x] >> 1
	},
}

var lShiftMap map[bool]func(uint16) = map[bool]func(uint16){
	true: func(opcode uint16) { // Cosmac compatible right shift function
		x := getOpcodeNibble(opcode, 1)
		y := getOpcodeNibble(opcode, 2)
		V[x] = V[y]
		V[0xF] = (V[x] & 0x1)
		V[x] = V[x] << 1
	},
	false: func(opcode uint16) { // Chip-48/Super-Chip compatible right shift function
		x := getOpcodeNibble(opcode, 1)
		V[0xF] = (V[x] & 0x1)
		V[x] = V[x] << 1
	},
}

var jumpMap map[bool]func(uint16) = map[bool]func(uint16){
	true: func(opcode uint16) { // Cosmac compatible JPb function
		nnn := opcode & 0x0FFF
		PC = nnn + uint16(V[0x0])
	},
	false: func(opcode uint16) { // Chip-48/Super-Chip compatible JPb function
		x := getOpcodeNibble(opcode, 1)
		nn := getOpcodeByte(opcode, 1)
		PC = uint16(nn) + uint16(V[x])
	},
}

var yCoordMap map[bool]func(byte, byte) int = map[bool]func(byte, byte) int{
	true: func(yStart byte, addedRows byte) int { // Vertical wrapping enabled
		return int((yStart + byte(addedRows)) % DH)
	},
	false: func(yStart byte, addedRows byte) int { // vertical wrapping disabled
		return int(yStart + addedRows)
	},
}

// Returns a byte of a chip8 opcode as defined by index. index may only be values 0 and 1. all other values are out of bounds
func getOpcodeByte(opcode uint16, index byte) byte {
	switch index {
	case 0:
		return byte((0xFF00 & opcode) >> 8)
	case 1:
		return byte((0x00FF & opcode))
	default:
		log.Fatal("Fatal: Attempt to access opcode byte is out of bounds at index", index)
	}
	return 0
}

// Returns a nibble of a chip8 opcode as defined by index. index may only be a value between 0 and 3. all other values are out of bounds.
func getOpcodeNibble(opcode uint16, index byte) byte {
	switch index {
	case 0:
		return byte((0xF000 & opcode) >> 12)
	case 1:
		return byte((0x0F00 & opcode) >> 8)
	case 2:
		return byte((0x00F0 & opcode) >> 4)
	case 3:
		return byte((0x000F & opcode))
	default:
		log.Fatal("Fatal: Attempt to access opcode nibble is out of bounds at index", index)
	}
	return 0
}

// CLS ...00E0: Clears the screen using the provided IO interface.
func CLS(inout io.IO) {
	px := inout.GetPixels()
	for row := range px {
		for col := range px[row] {
			inout.SetPixel(col, row, false)
		}
	}
	inout.Refresh()
}

// RET ...00EE: Pops the last memory address from the stack and updates the program counter with this address.
func RET() {
	PC, _ = STK.Pop()
}

// JP ...1nnn: Updates the program counter to the address specified in the last three nibbles of the opcode.
func JP(opcode uint16) {
	PC = opcode & 0x0FFF
}

// CALL ...2nnn: Pushes the current program counter value onto the stack and then updates the program counter to the address specified in the opcode.
func CALL(opcode uint16) {
	STK.Push(PC)
	PC = opcode & 0x0FFF
}

// SE3 ...3xkk: Compares the value at variable register V[x] with byte; if equal, increments the program counter by one instruction.
func SE3(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	if V[x] == kk {
		PC += 2
	}
}

// SNE4 ...4xkk: Compares the value at variable register V[x] with byte; if not equal, increments the program counter by one instruction.
func SNE4(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	if V[x] != kk {
		PC += 2
	}
}

// SE5 ...5xy0: Compares the values at variable registers V[x] and V[y]; if equal, increments the program counter by one instruction.
func SE5(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	if V[x] == V[y] {
		PC += 2
	}
}

// LD6 ...6xkk: Assigns byte kk to variable register V[x].
func LD6(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	V[x] = kk
}

// ADD7 ...7xkk: Adds byte kk to the value at variable register V[x] and updates V[x].
func ADD7(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	V[x] += kk
}

// LD8 ...8xy0: Copies the value from variable register V[y] to V[x].
func LD8(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[x] = V[y]
}

// OR ...8xy1: Performs a bitwise OR between V[x] and V[y], and stores the result in V[x].
func OR(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[x] = V[x] | V[y]
}

// AND ...8xy2: Performs a bitwise AND between V[x] and V[y], and stores the result in V[x].
func AND(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[x] = V[x] & V[y]
}

// XOR ...8xy3: Performs a bitwise XOR between V[x] and V[y], and stores the result in V[x].
func XOR(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[x] = V[x] ^ V[y]
}

// ADD8 ...8xy4: Adds V[y] to V[x]. If the addition results in overflow, sets V[F] to 1; otherwise, sets it to 0.
func ADD8(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	sum := V[x] + V[y]
	if sum < V[x] {
		V[0xF] = 1
	} else {
		V[0xF] = 0
	}
	V[x] = sum
}

// SUB ...8xy5: Subtracts V[y] from V[x]. If the subtraction results in underflow, sets V[F] to 0; otherwise, sets it to 1.
func SUB(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[0xF] = 1
	diff := V[x] - V[y]
	if diff > V[x] {
		V[0xF] = 0
	}
	V[x] = diff
}

// SHR ...8xy6: Shifts V[x] right by one bit. If the CosmacCompatibility configuration is true, copies V[y] to V[x] first.
func SHR(opcode uint16, conf Config) {
	rShiftMap[conf.CosmacCompatible](opcode)
}

// SUBN ...8xy7: Subtracts V[x] from V[y]. If the subtraction results in underflow, sets V[F] to 0; otherwise, sets it to 1.
func SUBN(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	V[0xF] = 1
	diff := V[y] - V[x]
	if diff > V[x] {
		V[0xF] = 0
	}
	V[x] = diff
}

// SHL ...8xyE: Shifts V[x] left by one bit. If the CosmacCompatibility configuration is true, copies V[y] to V[x] first.
func SHL(opcode uint16, conf Config) {
	lShiftMap[conf.CosmacCompatible](opcode)
}

// SNE9 ...9xy0: Compares the values at variable registers V[x] and V[y]; if not equal, increments the program counter by one instruction.
func SNE9(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	if V[x] != V[y] {
		PC += 2
	}
}

// LDa ...Annn: Loads the address specified in the last three nibbles of the opcode into the index register.
func LDa(opcode uint16) {
	I = opcode & 0x0FFF
}

// JPb ...Bnnn: Based on the configuration, either adds the value at V0 to the address and loads it into the program counter or uses the address offset by the value at VX.
func JPb(opcode uint16, conf Config) {
	jumpMap[conf.CosmacCompatible](opcode)
}

// RND ...Cxnn: Generates a random byte, performs a bitwise AND with nn, and stores the result in V[x].
func RND(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	nn := getOpcodeByte(opcode, 1)
	V[x] = byte(rand.Uint32()) & nn
}

// DRW ...Dxyn: XORs a sprite onto the screen at the coordinate (V[x], V[y]) that has a width of 8 pixels and a height of n pixels, based on the data starting at the address stored in I.
func DRW(opcode uint16, conf Config, inout io.IO) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	n := int(getOpcodeNibble(opcode, 3))
	if y >= byte(len(V)) || x >= byte(len(V)) {
		log.Fatalf("Fatal: out of bounds error in draw: X:%x Y:%x", x, y)
	}

	xStart := V[x] & (DW - 1)
	yStart := V[y] & (DH - 1)
	V[0xF] = 0
	for i := 0; i < n; i++ { // for each row in the sprite
		yCoord := yCoordMap[conf.VerticalWrapping](yStart, byte(i))
		if yCoord >= int(DH) {
			break
		}
		spriteRow := MEM[I+uint16(i)]
		processedBits := int(math.Min(float64(DW-xStart), 8))
		for j := processedBits - 1; j >= 0; j-- { // for each bit in the sprite (left to right)
			xCoord := int(xStart) + j
			dspPxlOn, err := inout.GetPixel(yCoord, xCoord)
			if err != nil {
				log.Fatal(err)
			}
			dspPxl := dataconverter.BoolToByte(dspPxlOn)
			spritePxl := (spriteRow >> byte(j)) & 1
			V[0xF] |= dspPxl & spritePxl
			newPxl := dataconverter.ByteToBool(spritePxl ^ dspPxl)
			err = inout.SetPixel(yCoord, xCoord, newPxl)
		}
	}
	inout.Refresh()
}
