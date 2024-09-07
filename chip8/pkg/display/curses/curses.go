package curses

import (
	"fmt"
	"log"

	"github.com/gbin/goncurses"
)

var screen goncurses.Window

func Initialize() {
	scr, err := goncurses.Init()
	if err != nil {
		log.Fatal("goncurses.init:", err)
	}
	screen = *scr

	goncurses.Echo(false)

	// Colors
	if !goncurses.HasColors() {
		fmt.Println("Terminal doesnt support colors")
	}
	goncurses.StartColor()
	goncurses.InitPair(1, int16(214), 0)
	screen.AttrOn(goncurses.Char(goncurses.A_BOLD))
	screen.ColorOn(1)

	// Clear screen
	screen.Clear()
	screen.Refresh()

	// Display test
	maxY, maxX := screen.MaxYX()
	screen.MovePrint(maxY/2, maxX/2, "Hello, NCurses!")
	screen.Refresh()
	screen.Move(0, 0)
	screen.Refresh()

	// wait for user input
	for {
		ch := screen.GetChar()
		if string(ch) == "q" {
			break
		}
	}

}

func Terminate() {
	goncurses.End()
}

func DemoColors() {
	stdscr, err := goncurses.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer goncurses.End()

	// Enable colors
	if !goncurses.HasColors() {
		fmt.Println("This terminal does not support colors.")
		return
	}
	goncurses.StartColor()

	// Print a grid of colors (16 colors per row, 16 rows total for 256 colors)
	rows := 16
	cols := 16
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			colorID := int16(i*cols + j)
			goncurses.InitPair(colorID, colorID, goncurses.C_BLACK) // Set color pair

			stdscr.ColorOn(colorID)        // Turn on the color for this number
			stdscr.Printf("%3d ", colorID) // Print the color ID, formatted to align
			stdscr.ColorOff(colorID)       // Turn off the color
		}
		stdscr.Print("\n") // Move to the next line for a new row
	}

	// Refresh the screen to show the grid
	stdscr.Refresh()

	// Wait for user input before exiting
	stdscr.GetChar()
}
