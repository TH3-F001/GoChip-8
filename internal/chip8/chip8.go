package chip8

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/TH3-F001/GoChip-8/chip8/internal/config"
	"github.com/TH3-F001/GoChip-8/chip8/internal/dataconverter"
	"github.com/TH3-F001/GoChip-8/chip8/internal/io"
	"github.com/TH3-F001/gotoolshed/stack"
)

// Chip8 ... Struct that holds all of Chip8's registers, timers, and state variables
type Chip8 struct {
	// Chip8.MEM ... A 4KB byte arrow for storing the chip-8s working memory
	MEM [4096]byte
	// Chip8.STK ... A 16 element array of 16-bit memory addresses. Used to store previous memory address before jumping or calling a subroutine
	STK *stack.Stack[uint16]
	// Chip8.V ... an array of 16 byte-long variable registers for storing general purpose data
	V [16]byte
	// Chip8.PC ... a 16-bit Program Counter that stores the index of the currently running instruction in memory
	PC uint16
	// Chip8.SP ... a 16-bit Stack pointer that im not sure i've used yet, or why id use it... recursion maybe
	SP uint16
	// Chip8.I ... a 16-bit index register. Used to point at locations in memory
	I uint16
	// DT ... A byte-long Delay timer. Decremented 60 times a second until reaching zero
	DT byte
	// ST ... A byte-long timer similar to DT. Decremented 60 times a second until reaching zero. gives off a beeping sound as long as timer isnt zero
	ST byte

	// dw ... Display Width: holds the max width of the display
	dw byte
	// dh ... Display Height: holds the max height of the display
	dh byte
	// inout ... the Chip's local reference to the io.IO object
	inout io.IO
	// ticker ... a 60Hz ticker for the delay and sound timers
	ticker *time.Ticker

	// RightShiftFunc ... A function pointer that is assigned on initialization based on whether the CosmacCompatible flag is true
	RightShiftFunc func(*Chip8, uint16)
	// LeftShiftFunc ... A function pointer that is assigned on initialization based on whether the CosmacCompatible flag is true
	LeftShiftFunc func(*Chip8, uint16)
	// JumpbFunc ... A function pointer that is assigned on initialization based on whether the CosmacCompatible flag is true
	JumpbFunc func(*Chip8, uint16)
	// YCoordFunc ... A function pointer that is assigned on initialization based on whether the CosmacCompatible flag is true
	YCoordFunc func(*Chip8, byte, byte) int
}

// #region ShiftFunc Implementations
func rightShiftCosmac(chip *Chip8, opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[y]
	chip.V[0xF] = (chip.V[x] & 0x1)
	chip.V[x] = chip.V[x] >> 1
}

func rightShiftSuper(chip *Chip8, opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	chip.V[0xF] = (chip.V[x] & 0x1)
	chip.V[x] = chip.V[x] >> 1
}

func leftShiftCosmac(chip *Chip8, opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[y]
	chip.V[0xF] = (chip.V[x] & 0x1)
	chip.V[x] = chip.V[x] << 1
}

func leftShiftSuper(chip *Chip8, opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	chip.V[0xF] = (chip.V[x] & 0x1)
	chip.V[x] = chip.V[x] << 1
}

//#endregion

// #region JumpbFunc Implementations
func jumpbCosmac(chip *Chip8, opcode uint16) {
	nnn := opcode & 0x0FFF
	chip.PC = nnn + uint16(chip.V[0x0])
}

func jumpbSuper(chip *Chip8, opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	nn := getOpcodeByte(opcode, 1)
	chip.PC = uint16(nn) + uint16(chip.V[x])
}

//#endregion

// #region YCoordFunc Implementations
func getYCoordWrapped(chip *Chip8, yStart byte, addedRows byte) int {
	return int((yStart + addedRows) % chip.dh)
}

func getYCoord(chip *Chip8, yStart byte, addedRows byte) int {
	return int(yStart + addedRows)
}

//#endregion

// #region Helper Functions

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

//#endregion

// New ... Maps configurable functions to the chip8's function pointers, and gives chip8 a local reference to an io.IO instance
func New(conf config.Config, inout io.IO, program, font []byte, displayHeight, displayWidth byte) *Chip8 {
	var chip Chip8
	if conf.CosmacCompatible {
		chip.RightShiftFunc = rightShiftCosmac
		chip.LeftShiftFunc = leftShiftCosmac
		chip.JumpbFunc = jumpbCosmac
	} else {
		chip.RightShiftFunc = rightShiftSuper
		chip.LeftShiftFunc = leftShiftSuper
		chip.JumpbFunc = jumpbSuper
	}
	if conf.VerticalWrapping {
		chip.YCoordFunc = getYCoordWrapped
	} else {
		chip.YCoordFunc = getYCoord
	}
	chip.inout = inout
	chip.STK = stack.New[uint16](32)

	copy(chip.MEM[0x50:], font)
	copy(chip.MEM[0x200:], program)
	chip.PC = 0x200
	chip.ST, chip.DT = 255, 255
	chip.dh = displayHeight
	chip.dw = displayWidth

	chip.ticker = time.NewTicker(time.Second / 60)
	return &chip
}

// Terminate ... Terminates hanging chip8 resources
func (chip *Chip8) Terminate() {
	chip.ticker.Stop()
}

// #region OpCodes

// CLS ...00E0: Clears the screen using the provided IO interface.
func (chip *Chip8) CLS() {
	px := chip.inout.GetPixels()
	for row := range px {
		for col := range px[row] {
			chip.inout.SetPixel(col, row, false)
		}
	}
	chip.inout.Refresh()
}

// RET ...00EE: Pops the last memory address from the stack and updates the program counter with this address.
func (chip *Chip8) RET() {
	address, _ := chip.STK.Pop()
	chip.PC = address
}

// JP ...1nnn: Updates the program counter to the address specified in the last three nibbles of the opcode.
func (chip *Chip8) JP(opcode uint16) {
	chip.PC = opcode & 0x0FFF
}

// CALL ...2nnn: Pushes the current program counter value onto the stack and then updates the program counter to the address specified in the opcode.
func (chip *Chip8) CALL(opcode uint16) {
	chip.STK.Push(chip.PC)
	chip.PC = opcode & 0x0FFF
}

// SE3 ...3xkk: Compares the value at variable register V[x] with byte; if equal, increments the program counter by one instruction.
func (chip *Chip8) SE3(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	if chip.V[x] == kk {
		chip.PC += 2
	}
}

// SNE4 ...4xkk: Compares the value at variable register V[x] with byte; if not equal, increments the program counter by one instruction.
func (chip *Chip8) SNE4(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	if chip.V[x] != kk {
		chip.PC += 2
	}
}

// SE5 ...5xy0: Compares the values at variable registers V[x] and V[y]; if equal, increments the program counter by one instruction.
func (chip *Chip8) SE5(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	if chip.V[x] == chip.V[y] {
		chip.PC += 2
	}
}

// LD6 ...6xkk: Assigns byte kk to variable register V[x].
func (chip *Chip8) LD6(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	chip.V[x] = kk
}

// ADD7 ...7xkk: Adds byte kk to the value at variable register V[x] and updates V[x].
func (chip *Chip8) ADD7(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	kk := getOpcodeByte(opcode, 1)
	chip.V[x] += kk
}

// LD8 ...8xy0: Copies the value from variable register V[y] to V[x].
func (chip *Chip8) LD8(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[y]
}

// OR ...8xy1: Performs a bitwise OR between V[x] and V[y], and stores the result in V[x].
func (chip *Chip8) OR(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[x] | chip.V[y]
}

// AND ...8xy2: Performs a bitwise AND between V[x] and V[y], and stores the result in V[x].
func (chip *Chip8) AND(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[x] & chip.V[y]
}

// XOR ...8xy3: Performs a bitwise XOR between V[x] and V[y], and stores the result in V[x].
func (chip *Chip8) XOR(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	chip.V[x] = chip.V[x] ^ chip.V[y]
}

// ADD8 ...8xy4: Adds V[y] to V[x]. If the addition results in overflow, sets V[F] to 1; otherwise, sets it to 0.
func (chip *Chip8) ADD8(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	sum := chip.V[x] + chip.V[y]
	if sum < chip.V[x] {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0
	}
	chip.V[x] = sum
}

// SUB ...8xy5: Subtracts V[y] from V[x]. If the subtraction results in underflow, sets V[F] to 0; otherwise, sets it to 1.
func (chip *Chip8) SUB(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	diff := chip.V[x] - chip.V[y]
	if diff > chip.V[x] {
		chip.V[0xF] = 0
	} else {
		chip.V[0xF] = 1
	}
	chip.V[x] = diff
}

// SHR ...8xy6: Shifts V[x] right by one bit. If the CosmacCompatibility configuration is true, copies V[y] to V[x] first.
func (chip *Chip8) SHR(opcode uint16) {
	chip.RightShiftFunc(chip, opcode)
}

// SUBN ...8xy7: Subtracts V[x] from V[y]. If the subtraction results in underflow, sets V[F] to 0; otherwise, sets it to 1.
func (chip *Chip8) SUBN(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	diff := chip.V[y] - chip.V[x]
	if diff > chip.V[x] {
		chip.V[0xF] = 0
	} else {
		chip.V[0xF] = 1
	}
	chip.V[x] = diff
}

// SHL ...8xyE: Shifts V[x] left by one bit. If the CosmacCompatibility configuration is true, copies V[y] to V[x] first.
func (chip *Chip8) SHL(opcode uint16) {
	chip.LeftShiftFunc(chip, opcode)
}

// SNE9 ...9xy0: Compares the values at variable registers V[x] and V[y]; if not equal, increments the program counter by one instruction.
func (chip *Chip8) SNE9(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	if chip.V[x] != chip.V[y] {
		chip.PC += 2
	}
}

// LDa ...Annn: Loads the address specified in the last three nibbles of the opcode into the index register.
func (chip *Chip8) LDa(opcode uint16) {
	chip.I = opcode & 0x0FFF
}

// JPb ...Bnnn: Based on the configuration, either adds the value at V0 to the address and loads it into the program counter or uses the address offset by the value at VX.
func (chip *Chip8) JPb(opcode uint16) {
	chip.JumpbFunc(chip, opcode)
}

// RND ...Cxnn: Generates a random byte, performs a bitwise AND with nn, and stores the result in V[x].
func (chip *Chip8) RND(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	nn := getOpcodeByte(opcode, 1)
	chip.V[x] = byte(rand.Uint32()) & nn
}

// DRW ...Dxyn: XORs a sprite onto the screen at the coordinate (V[x], V[y]) that has a width of 8 pixels and a height of n pixels, based on the data starting at the address stored in I.
func (chip *Chip8) DRW(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	y := getOpcodeNibble(opcode, 2)
	n := int(getOpcodeNibble(opcode, 3))
	if y >= byte(len(chip.V)) || x >= byte(len(chip.V)) {
		log.Fatalf("Fatal: out of bounds error in draw: X:%x Y:%x", x, y)
	}

	xStart := chip.V[x] & (chip.dw - 1)
	yStart := chip.V[y] & (chip.dh - 1)
	chip.V[0xF] = 0
	for i := 0; i < n; i++ { // for each row in the sprite
		yCoord := chip.YCoordFunc(chip, yStart, byte(i))
		spriteRow := chip.MEM[chip.I+uint16(i)]
		for j := 0; j < 8; j++ { // for each bit in the sprite (left to right)
			xCoord := (int(xStart) + j) % int(chip.dw)
			spritePxl := (spriteRow >> (7 - j)) & 1
			dspPxlOn, err := chip.inout.GetPixel(yCoord, xCoord)
			if err != nil {
				log.Fatal(err)
			}
			dspPxl := dataconverter.BoolToByte(dspPxlOn)
			newPxl := dataconverter.ByteToBool(spritePxl ^ dspPxl)
			err = chip.inout.SetPixel(yCoord, xCoord, newPxl)
			chip.V[0xF] |= dspPxl & spritePxl
		}
	}
	chip.inout.Refresh()
}

// SKP ... Ex9E: Skips the next instruction if the key with value V[x] is pressed
func (chip *Chip8) SKP(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	key, _ := chip.inout.ListenNow()
	if key == chip.V[x] {
		chip.PC += 2
	}
}

// SKNP ... ExA1: Skips the next instruction if the key with the value V[x] is not pressed
func (chip *Chip8) SKNP(opcode uint16) {
	x := getOpcodeNibble(opcode, 1)
	key, _ := chip.inout.ListenNow()
	if key != chip.V[x] {
		chip.PC += 2
	}
}

// LDf ...

//#endregion

// MainLoop ... The main switch statement for the Chip8 interpreter. Fetches the opcode that PC is currently pointing to, parses it, and executes it
func (chip *Chip8) MainLoop() {
	// Fetch current OpCode
	opcode := uint16(chip.MEM[chip.PC])<<8 | uint16(chip.MEM[chip.PC+1])
	chip.PC += 2
	firstNibble := getOpcodeNibble(opcode, 0)

	select {
	case <-chip.ticker.C:
		if chip.DT > 0 {
			chip.DT--
		}
		if chip.ST > 0 {
			chip.ST--
			// chip.inout.Beep()
		}
	}

	// Decode and Execute
	switch firstNibble {
	case 0x0:
		lastByte := getOpcodeByte(opcode, 1)
		switch lastByte {
		case 0xE0:
			chip.CLS() // 00E0 - CLS
		case 0xEE:
			chip.RET() // 00EE - RET
		}
	case 0x1:
		chip.JP(opcode) // 1nnn - JP(addr)
	case 0x2:
		chip.CALL(opcode) // 2nnn - CALL(addr)
	case 0x3:
		chip.SE3(opcode) // 3xkk - SE(Vx, byte)
	case 0x4:
		chip.SNE4(opcode) // 4xkk - SNE(Vx, byte)
	case 0x5:
		chip.SE5(opcode) // 5xy0 - SE(Vx, Vy)
	case 0x6:
		chip.LD6(opcode) // 6xkk - LD(Vx, byte)
	case 0x7:
		chip.ADD7(opcode) // 7xkk - ADD(Vx, byte)
	case 0x8:
		lastNibble := getOpcodeNibble(opcode, 3)
		switch lastNibble {
		case 0x0:
			chip.LD8(opcode) // 8xy0 - LD(Vx, Vy)
		case 0x1:
			chip.OR(opcode) // 8xy1 - OR(Vx, Vy)
		case 0x2:
			chip.AND(opcode) // 8xy2 - AND(Vx, Vy)
		case 0x3:
			chip.XOR(opcode) // 8xy3 - XOR(Vx, Vy)
		case 0x4:
			chip.ADD8(opcode) // 8xy4 - ADD(Vx, Vy)
		case 0x5:
			chip.SUB(opcode) // 8xy5 - SUB(Vx, Vy)
		case 0x6:
			chip.SHR(opcode) // 8xy6 - SHR(Vx Vy)
		case 0x7:
			chip.SUBN(opcode) // 8xy7 - SUBN(Vx, Vy)
		case 0xE:
			chip.SHL(opcode) // 8xyE - SHL(Vx Vy)
		}
	case 0x9:
		chip.SNE9(opcode) // 9xy0 - SE (Vx, Vy)
	case 0xA:
		chip.LDa(opcode) // Annn - LD (I, addr) (Load the value nnn into register I)
	case 0xB: // SuperChipCompatible: Bxnn - JP(Vx, addr)
		chip.JPb(opcode) // Cosmac Compatible: Bnnn - JP (V0, addr)
	case 0xC:
		chip.RND(opcode) // Cxnn - RND(Vx, byte)
	case 0xD: // Dxyn - DRW (Vx, Vy, nibble)
		chip.DRW(opcode)
	}
}
