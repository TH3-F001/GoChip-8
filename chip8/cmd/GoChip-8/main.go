package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/display"
	// "github.com/TH3-F001/GoChip-8/pkg/display/curses"
)

type Config struct {
	DisplayType           string
	InputType             string
	FgColor               int32
	BgColor               int32
	InstructionsPerSecond int32
}

var memory [4096]byte = [4096]byte{}
var stack [16]uint16 = [16]uint16{}
var v [16]byte = [16]byte{} // Variable Registers
var pc uint16               // Program Counter
var sp uint16               // Stack Pointer
var ir uint16               // Index Register
var dt byte                 // Delay Timer
var st byte                 // Sound Timer

var conf Config
var screen display.Display

func getConfigPath() string {
	configPath, exists := os.LookupEnv("CHIP_8_CONF_PATH")
	if !exists {
		configDir, err := os.UserConfigDir()
		if err == nil {
			configPath = filepath.Join(configDir, "GoChip-8", "chip8.toml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				configPath, err = filepath.Abs("../../config/chip8")
				if err != nil {
					log.Fatal("Fatal: Cannot find chip8.toml: ", err)
				}
			}
		} else {
			configPath, err = filepath.Abs("../../config/chip8")
			if err != nil {
				log.Fatal("Fatal: Cannot find chip8.toml: ", err)
			}
		}
	}
	fmt.Println("\tFound configuration at:", configPath)
	return configPath
}

func loadConfig(path string) {
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		conf = Config{
			DisplayType:           "ansi",
			InputType:             "sdl",
			FgColor:               32,
			BgColor:               30,
			InstructionsPerSecond: 700,
		}
	}

}

func main() {
	fmt.Println("Initializing GoChip-8...")

	fmt.Println("Loading Config...")
	confPath := getConfigPath()
	loadConfig(confPath)

	fmt.Printf("\t%v\n", conf)

	// // Create Display
	// screen, err := ansi.NewDisplay(64, 32, 30, 33)
	// if err != nil {
	// 	log.Fatal("Fatal: Failed to create a new display: ", err)
	// }

	// for {

	// }

}
