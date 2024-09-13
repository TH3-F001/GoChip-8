package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/sdlio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/tcellio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/vanillaio"
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
}

var memory [4096]byte = [4096]byte{}
var stack [16]uint16 = [16]uint16{}
var v [16]byte = [16]byte{} // Variable Registers
var pc uint16               // Program Counter
var sp uint16               // Stack Pointer
var ir uint16               // Index Register
var dt byte                 // Delay Timer
var st byte                 // Sound Timer

const width byte = 64
const height byte = 32

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

	copy(memory[0x50:], fontBytes)
}

func createIo(conf Config) (io.IO, error) {
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

// #region Instructions
// op-code 00E0 - clears the screen

//#endregion


func mainLoop(inout io.IO, conf Config) {
	//Fetch
	// Get the first byte, shift it to the left of the double, and then drop in the next byte with an or
	// see docs/visualizations/opcode-fetch.png
	opCode := uint16(memory[pc])<<8 | uint16(memory[pc+1])
	pc += 2

	var nibbles [4]uint16 = [4]uint16{
		(uint16(0xF000) & opCode) >> 12,
		(uint16(0x0F00) & opCode) >> 8,
		(uint16(0x00F0) & opCode) >> 4,
		(uint16(0x000F) & opCode),
	}

	// Decode and Execute
	switch nibbles[0] {
	case 0x0:
		if opCode == 0x00E0 {	// CLS
			px := *(inout.GetPixels())
			for i := range px {
				for j := range (px)[i] {
					inout.SetPixel(j, i, false)
				}
			}
			inout.Refresh()
		}
	case 0x1:	// JMP
		pc = opCode & 0x0FFF
	}

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

		
		inout.Refresh()
		input, err := inout.Listen()
		if err != nil && input == 255 {
			os.Exit(0)
		}
	}

	for pixels := range *inout.GetPixels() {
		fmt.Println(pixels)
	}

}
