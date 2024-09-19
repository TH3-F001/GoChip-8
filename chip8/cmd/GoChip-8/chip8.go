package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io"

	// "github.com/TH3-F001/GoChip-8/chip8/pkg/io/sdlio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/tcellio"
	// "github.com/TH3-F001/GoChip-8/chip8/pkg/io/vanillaio"
	"github.com/TH3-F001/gotoolshed/stack"
)

//go:embed config/chip8.toml
var embeddedConf embed.FS

//go:embed fonts/*
var fonts embed.FS

//go:embed demo/*
var demoProgs embed.FS

// Config ... Struct for storing config data in memory. Program config is
// pulled from chip8.toml by default. But can be overridden with command line flags
type Config struct {
	IOType                string
	DefaultFont           string
	FgColor               uint32
	BgColor               uint32
	InstructionsPerSecond uint32
	SuperChip             bool
	CosmacCompatible      bool
	VerticalWrapping      bool
	ProgramPath           string
}

// MEM ... A 4KB byte arrow for storing the chip-8s working memory
var MEM [4096]byte = [4096]byte{}

// STK ... A 16 element array of 16-bit memory addresses. Used to store previous memory address before jumping or calling a subroutine
var STK *stack.Stack[uint16] = stack.New[uint16](16)

// V ... an array of 16 byte-long variable registers for storing general purpose data
var V [16]byte = [16]byte{}

// PC ... a 16-bit Program Counter that stores the index of the currently running instruction in memory
var PC uint16

// SP ... a 16-bit Stack pointer that im not sure i've used yet, or why id use it... recursion maybe
var SP uint16

// I ... a 16-bit index register. Used to point at locations in memory
var I uint16

// DT ... A byte-long Delay timer. Decremented 60 times a second until reaching zero
var DT byte

// ST ... A byte-long timer similar to DT. Decremented 60 times a second until reaching zero. gives off a beeping sound as long as timer isnt zero
var ST byte // Sound Timer

// DW ... Display Width: holds the max width of the display
var DW byte

// DH ... Display Height: holds the max height of the display
var DH byte // Display Height

// terminationCh ... A go channel used by io.ListenForTerminate to listen for user termination
var terminationCh chan bool = make(chan bool)

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
	var rawFontData []byte
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
	var dw int
	var dh int
	if conf.SuperChip {
		dw = 128
		dh = 64
	} else {
		dw = 64
		dh = 32
	}
	DH = byte(dh)
	DW = byte(dw)
	var io io.IO
	var err error
	switch conf.IOType {
	case "tcellio", "tcell", "tui":
		io, err = tcellio.New(dh, dw, conf.FgColor, conf.BgColor)
		if err != nil {
			log.Fatal("Fatal: Failed to Create new TcellIO instance: ", err)
		}
	// case "vanilla", "terminal", "term":
	// 	io, err = vanillaio.New(dw, dh, conf.FgColor, conf.BgColor)
	// 	if err != nil {
	// 		log.Fatal("Fatal: Failed to Create new VanillaIO instance: ", err)
	// 	}
	// case "sdl", "graphical", "gui":
	// 	io, err = sdlio.New(dw, dh, conf.FgColor, conf.BgColor)
	// 	if err != nil {
	// 		log.Fatal("Fatal: Failed to Create new VanillaIO instance: ", err)
	// 	}

	default:
		log.Fatal("Fatal: Failed to Create new IO instance: Invalid ioType: ", conf.IOType)
	}
	io.ListenForTermination(terminationCh)
	return io, nil
}

//#endregion

func loadProgram(conf Config) {
	var rawProgramData []byte
	var err error
	if conf.ProgramPath == "" {
		rawProgramData, err = demoProgs.ReadFile("demo/heart_monitor.ch8")
		if err != nil {
			log.Fatal("Fatal: Failed to load default program file: ", err)
		}
	}
	copy(MEM[0x200:], rawProgramData)
	PC = 0x200
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

	fmt.Println("\tLoading Program ByteCode Into Memory...")
	loadProgram(conf)
	fmt.Println("\t\tProgram Loaded.")

	fmt.Println("\tInitializing I/O...")
	inout, err := createIo(conf)
	if err != nil {
		fmt.Println("\t\tIO Initialized.")
		log.Fatal("\t\tFatal: Failed to create IO instance")
	}
	fmt.Println("\t\tIO Initialized.")

	// Main Loop
	for {
		select {
		case <-terminationCh:
			inout.Terminate()
			return
		default:
			time.Sleep(delay)
			// Fetch (see docs/visualizations/opcode-fetch.png)
			opcode := uint16(MEM[PC])<<8 | uint16(MEM[PC+1])
			PC += 2
			firstNibble := getOpcodeNibble(opcode, 0)

			// Decode and Execute
			switch firstNibble {
			case 0x0:
				lastByte := getOpcodeByte(opcode, 1)
				switch lastByte {
				case 0xE0:
					CLS(inout) // 00E0 - CLS
				case 0xEE:
					RET() // 00EE - RET
				}
			case 0x1:
				JP(opcode) // 1nnn - JP(addr)
			case 0x2:
				CALL(opcode) // 2nnn - CALL(addr)
			case 0x3:
				SE3(opcode) // 3xkk - SE(Vx, byte)
			case 0x4:
				SNE4(opcode) // 4xkk - SNE(Vx, byte)
			case 0x5:
				SE5(opcode) // 5xy0 - SE(Vx, Vy)
			case 0x6:
				LD6(opcode) // 6xkk - LD(Vx, byte)
			case 0x7:
				ADD7(opcode) // 7xkk - ADD(Vx, byte)
			case 0x8:
				lastNibble := getOpcodeNibble(opcode, 3)
				switch lastNibble {
				case 0x0:
					LD8(opcode) // 8xy0 - LD(Vx, Vy)
				case 0x1:
					OR(opcode) // 8xy1 - OR(Vx, Vy)
				case 0x2:
					AND(opcode) // 8xy2 - AND(Vx, Vy)
				case 0x3:
					XOR(opcode) // 8xy3 - XOR(Vx, Vy)
				case 0x4:
					ADD8(opcode) // 8xy4 - ADD(Vx, Vy)
				case 0x5:
					SUB(opcode) // 8xy5 - SUB(Vx, Vy)
				case 0x6:
					SHR(opcode, conf) // 8xy6 - SHR(Vx Vy)
				case 0x7:
					SUBN(opcode) // 8xy7 - SUBN(Vx, Vy)
				case 0xE:
					SHL(opcode, conf) // 8xyE - SHL(Vx Vy)
				}
			case 0x9:
				SNE9(opcode) // 9xy0 - SE (Vx, Vy)
			case 0xA:
				LDa(opcode) // Annn - LD (I, addr) (Load the value nnn into register I)
			case 0xB: // SuperChipCompatible: Bxnn - JP(Vx, addr)
				JPb(opcode, conf) // Cosmac Compatible: Bnnn - JP (V0, addr)
			case 0xC:
				RND(opcode) // Cxnn - RND(Vx, byte)
			case 0xD: // Dxyn - DRW (Vx, Vy, nibble)
				DRW(opcode, conf, inout)
			}
		}
	}
}
