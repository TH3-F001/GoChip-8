package main

import (
	"embed"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/sdlio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/tcellio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/vanillaio"
	"github.com/TH3-F001/gotoolshed/stack"
)

//go:embed config/chip8.toml
var embeddedConf embed.FS

//go:embed fonts/*
var fonts embed.FS

type Config struct {
	IOType                string
	DefaultFont           string
	FgColor               uint32
	BgColor               uint32
	InstructionsPerSecond uint32
	SuperChip             bool
	CosmacCompatible      bool
}

var MEM [4096]byte = [4096]byte{}
var STK *stack.Stack[uint16] = stack.New[uint16](16)
var V [16]byte = [16]byte{} // Variable Registers
var PC uint16               // Program Counter
var SP uint16               // Stack Pointer
var I uint16                // Index Register
var DT byte                 // Delay Timer
var ST byte                 // Sound Timer

// #region Configuration
func getConfigPath() string {
	configPath := ""

	if path, exists := os.LookupEnv("CHIP_8_CONF_PATH"); exists {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			return configPath
		}
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("Fatal: Cannot find user config directory: ", err)
	}

	configPath = filepath.Join(configDir, "GoChip-8", "chip8.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		//  If all else fails, fall back on the embedded config file and build a new config dir/file
		buildConfigDirectory()
	}
	return configPath
}

func buildConfigDirectory() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("Fatal: Cannot find user config directory: ", err)
	}
	configDir = filepath.Join(configDir, "GoChip-8")
	configPath := filepath.Join(configDir, "chip8.toml")
	defaultConfig, err := embeddedConf.ReadFile("config/chip8.toml")
	if err != nil {
		log.Fatal("Fatal: Failed to load embedded chip8.toml file")
	}

	if err = os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		log.Fatal("Fatal: Failed to create directory for new config file")
	}

	if err = os.WriteFile(configPath, defaultConfig, 0755); err != nil {
		log.Fatal("Fatal: Failed to create new config file")
	}

	return configPath
}

func loadConfig(path string) Config {
	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		path = buildConfigDirectory()
		if _, err := toml.DecodeFile(path, &conf); err != nil {
			log.Fatal("Fatal: failed to load chip8.toml: ", err)
		}
	}
	return conf
}

//#endregion

// #region Initialization
func loadDefaultFont(config Config) {
	rawFontData := make([]byte, 0)
	var err error = nil
	switch config.DefaultFont {
	case "chip48":
		rawFontData, err = fonts.ReadFile("fonts/chip48font.txt")
	case "cosmac":
		rawFontData, err = fonts.ReadFile("fonts/cosmacvipfont.txt")
	case "dream":
		rawFontData, err = fonts.ReadFile("fonts/dream6800font.txt")
	case "eti":
		rawFontData, err = fonts.ReadFile("fonts/eti660font.txt")
	default:
		rawFontData, err = fonts.ReadFile("fonts/chip48font.txt")
	}
	if err != nil {
		log.Fatal("Fatal: Failed to load default font file: ", err)
	}

	fontString := string(rawFontData)
	fontStringArr := make([]string, 0)

	segments := strings.Split(fontString, "0x")
	for _, seg := range segments {
		lines := strings.Split(seg, "\n")
		for _, line := range lines {
			if !strings.Contains(line, ":") && line != "" {
				str := strings.TrimSuffix(strings.TrimPrefix(line, "0b"), "\r")
				fontStringArr = append(fontStringArr, str)
			}
		}
	}

	fontBytes := make([]byte, 0)
	for _, str := range fontStringArr {
		val, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			log.Fatal("Fatal: couldnt convert font to binary: ", err)
		}
		fontBytes = append(fontBytes, byte(val))
	}

	copy(MEM[0x50:], fontBytes)
}

func createIo(conf Config) (io.IO, error) {
	var width, height byte
	if conf.SuperChip {
		width = byte(128)
		height = byte(64)
	} else {
		width = byte(64)
		height = byte(32)
	}
	var io io.IO
	var err error
	switch conf.IOType {
	case "tcellio", "tcell", "tui":
		io, err = tcellio.New(width, height, conf.FgColor, conf.BgColor)
		if err != nil {
			log.Fatal("Fatal: Failed to Create new TcellIO instance", err)
		}
	case "vanilla", "terminal", "term":
		io, err = vanillaio.New(width, height, conf.FgColor, conf.BgColor)
		if err != nil {
			log.Fatal("Fatal: Failed to Create new VanillaIO instance", err)
		}
	case "sdl", "graphical", "gui":
		io, err = sdlio.New(width, height, conf.FgColor, conf.BgColor)
		if err != nil {
			log.Fatal("Fatal: Failed to Create new VanillaIO instance", err)
		}

	default:
		log.Fatal("Fatal: Failed to Create new IO instance: Invalid ioType: ", conf.IOType)
	}
	return io, nil
}

//#endregion

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
		return byte((0x0F00 & opcode) >> 4)
	case 3:
		return byte((0x000F & opcode))
	default:
		log.Fatal("Fatal: Attempt to access opcode nibble is out of bounds at index", index)
	}
	return 0
}

func main() {
	var inout io.IO
	defer func() {
		inout.Terminate()
	}()
	fmt.Println("Initializing GoChip-8...")

	fmt.Println("\tLoading Config...")
	confPath := getConfigPath()
	fmt.Println("\t\tFound configuration at:", confPath)
	conf := loadConfig(confPath)
	fmt.Println("\t\tSetting instruction Delay...")
	delay := time.Second / time.Duration(conf.InstructionsPerSecond)
	fmt.Println("\t\tConfig Loaded.")

	fmt.Println("\tLoading Font...")
	loadDefaultFont(conf)
	fmt.Println("\t\tFont Loaded.")

	fmt.Println("Initializing I/O...")
	inout, err := createIo(conf)
	if err != nil {
		log.Fatal("Fatal: Failed to create IO instance")
	}

	// Main Loop
	for {
		// Delay execution to mimic instructions per second
		time.Sleep(delay)

		// Fetch (see docs/visualizations/opcode-fetch.png)
		opcode := uint16(MEM[PC])<<8 | uint16(MEM[PC+1])
		PC += 2
		firstNibble := getOpcodeNibble(opcode, 0)

		// Decode and Execute
		switch firstNibble {
		case 0x0:
			lastByte := getOpcodeByte(opcode, 1)
			if lastByte == 0xE0 { // 00E0 - CLS (clear screen)
				px := *(inout.GetPixels())
				for i := range px {
					for j := range (px)[i] {
						inout.SetPixel(j, i, false)
					}
				}
				inout.Refresh()
			} else if lastByte == 0xEE { // 00EE - RET (pop the top value from stack and set pc to that value )
				PC, _ = STK.Pop()
			}

		case 0x1: // 1nnn - JP addr (move pc to address nnn)
			PC = opcode & 0x0FFF

		case 0x2: // 2nnn - CALL addr (save pc to stack then set it to nnn)
			STK.Push(PC)
			PC = opcode & 0x0FFF

		case 0x3: // 3xnn - SE Vx, byte (skip next instruction if V[x] is equal to nn)
			x := getOpcodeNibble(opcode, 1)
			secondByte := getOpcodeByte(opcode, 1)
			if V[x] == secondByte {
				PC += 2
			}

		case 0x4: // 4xnn - SNE Vx, byte (skip next instruction if V[x] is not equal to nn)
			x := getOpcodeNibble(opcode, 1)
			secondByte := getOpcodeByte(opcode, 1)
			if V[x] != secondByte {
				PC += 2
			}

		case 0x5: // 5xy0 - SE Vx, Vy (skip next instruction if V[x] is equal to V[y])
			x := getOpcodeNibble(opcode, 1)
			y := getOpcodeNibble(opcode, 2)
			if V[x] == V[y] {
				PC += 2
			}

		case 0x6: // 6xnn - LD Vx, byte (load the value nn into V[x])
			x := getOpcodeNibble(opcode, 1)
			val := getOpcodeByte(opcode, 1)
			V[x] = val

		case 0x7: // 7xnn - ADD Vx, byte (adds the value of nn to value of V[x] and stores it back in to V[x])
			x := getOpcodeNibble(opcode, 1)
			val := getOpcodeByte(opcode, 1)
			V[x] += val

		case 0x8:
			lastNibble := getOpcodeNibble(opcode, 3)
			switch lastNibble {
			case 0x0: // 8xy0 - LD Vx, Vy (loads the value of V[y] into V[x])
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[x] = V[y]

			case 0x1: // 8xy1 - OR Vx, Vy (sets V[x] to the result of a binary OR between V[x] and V[y])
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[x] = V[x] | V[y]

			case 0x2: // 8xy2 - AND Vx, Vy (sets V[x] to the result of a binary AND between V[x] and V[y])
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[x] = V[x] & V[y]

			case 0x3: // 8xy3 - XOR Vx, Vy (sets V[x] to the result of a binary XOR between V[x] and V[y])
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[x] = V[x] ^ V[y]

			case 0x4: // 8xy4 - ADD Vx, Vy (sets V[x] to the value of V[x] + V[y], and sets V[F] to 1 if theres an overflow, else sets it to zero)
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				sum := V[x] + V[y]
				if sum < V[x] {
					V[0xF] = 1
				} else {
					V[0xF] = 0
				}
				V[x] = sum

			case 0x5: // 8xy5 - SUB Vx, Vy (sets V[x] to V[x] - V[y]. V[F] is set to 0 if an underflow occured, else 1)
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[0xF] = 1
				diff := V[x] - V[y]
				if diff > V[x] {
					V[0xF] = 0
				}
				V[x] = diff

			case 0x6: // 8xy6 - SHR Vx {, Vy} (if cosmac compatible copy v[y] to v[x]. regardless. shift v[x] right by one bit)
				x := getOpcodeNibble(opcode, 1)
				if conf.CosmacCompatible && !conf.SuperChip {
					y := getOpcodeNibble(opcode, 2)
					V[x] = V[y]
				}
				V[0xF] = (V[x] & 0x1)
				V[x] = V[x] >> 1

			case 0x7: // 8xy7 - SUB Vx, Vy (sets V[x] to V[y] - V[x]. V[F] is set to 0 if an underflow occured, else 1)
				x := getOpcodeNibble(opcode, 1)
				y := getOpcodeNibble(opcode, 2)
				V[0xF] = 1
				diff := V[y] - V[x]
				if diff > V[x] {
					V[0xF] = 0
				}
				V[x] = diff

			case 0xE: // 8xyE - SHL Vx {, Vy} (if cosmac compatible copy v[y] to v[x]. regardless. shift v[x] left by one bit)
				x := getOpcodeNibble(opcode, 1)
				if conf.CosmacCompatible && !conf.SuperChip {
					y := getOpcodeNibble(opcode, 2)
					V[x] = V[y]
				}
				V[0xF] = (V[x] & 0x80) >> 7
				V[x] = V[x] << 1
			}
		case 0x9: // 9xy0 - SE Vx, Vy (skip next instruction if V[x] is not equal to V[y])
			x := getOpcodeNibble(opcode, 1)
			y := getOpcodeNibble(opcode, 2)
			if V[x] != V[y] {
				PC += 2
			}

		case 0xA: // Annn - LD I, addr (Load the value nnn into register I)
			I = opcode & 0x0FFF

		case 0xB: // Bnnn/Bxnn - JP V0/Vx, nnn/nn (set PC to nn/nnn plus the value of V0/Vx)
			if conf.CosmacCompatible && !conf.SuperChip {
				nnn := opcode & 0x0FFF
				PC = nnn + uint16(V[0x0])
			} else {
				x := getOpcodeNibble(opcode, 1)
				nn := getOpcodeByte(opcode, 1)
				PC = uint16(nn) + uint16(V[x])
			}

		case 0xC: // Cxnn - RND Vx, byte (Set V[x] to "random byte & nn")
			x := getOpcodeNibble(opcode, 1)
			nn := getOpcodeByte(opcode, 1)
			V[x] = byte(rand.Uint32()) & nn

		case 0xD: // Dxyn - DRW Vx, Vy, nibble (big instruction)
			x := getOpcodeNibble(opcode, 1)
			y := getOpcodeNibble(opcode, 2)
			n := getOpcodeNibble(opcode, 3)

			spriteStart := MEM[I]
			spriteEnd := MEM[I+n]

		}

		// inout.Refresh()
		// input, err := inout.Listen()
		// if err != nil && input == 255 {
		// 	os.Exit(0)
		// }
	}

}
