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
	"github.com/TH3-F001/GoChip-8/chip8/pkg/chip8"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/config"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io"

	// "github.com/TH3-F001/GoChip-8/chip8/pkg/io/sdlio"
	"github.com/TH3-F001/GoChip-8/chip8/pkg/io/tcellio"
	// "github.com/TH3-F001/GoChip-8/chip8/pkg/io/vanillaio"
)

//go:embed config/chip8.toml
var embeddedConf embed.FS

//go:embed fonts/*
var fonts embed.FS

//go:embed demo/*
var demoProgs embed.FS

// terminationCh ... A go channel used by io.ListenForTerminate() to listen for user termination
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

func loadConfig(path string) config.Config {
	var conf config.Config
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
func getDefaultFont(config config.Config) []byte {
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

	return fontBytes
}

func getProgram(conf config.Config) []byte {
	var rawProgramData []byte
	var err error
	if conf.ProgramPath == "" {
		rawProgramData, err = demoProgs.ReadFile("demo/heart_monitor.ch8")
		if err != nil {
			log.Fatal("Fatal: Failed to load default program file: ", err)
		}
	}
	return rawProgramData
}

func createIo(conf config.Config) (io.IO, byte, byte, error) {
	var dw int
	var dh int
	if conf.CosmacCompatible {
		dw = 64
		dh = 32
	} else {
		dw = 128
		dh = 64
	}

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
	return io, byte(dh), byte(dw), nil
}

//#endregion

func main() {
	var inout io.IO

	defer func() {
		fmt.Println("C U Next Time!")
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

	fmt.Println("\tInitializing I/O...")
	inout, dh, dw,  err := createIo(conf)
	if err != nil {
		log.Fatal("\t\tFatal: Failed to create IO instance")
	}
	fmt.Println("\t\tIO Initialized.")

	fmt.Println("\tInitializing Chip Instance...")
	font := getDefaultFont(conf)
	program := getProgram(conf)
	chip := chip8.New(conf, inout, program, font, dh, dw)

	active := true
	// Main Loop
	for {
		select {
		case <-terminationCh:
			inout.Terminate()
			return
		default:
			if active {
				time.Sleep(delay)
				chip.MainLoop()
			}
		}
	}
}
