package curses

import (
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
	if goncurses.HasColors() {
		goncurses.StartColor()
	}
	goncurses.InitPair(1, int16(goncurses.C_YELLOW), goncurses.C_BLACK)
	screen.AttrOn(goncurses.Char(goncurses.A_BOLD))

	// Clear screen
	screen.Clear()
	screen.Refresh()

	// Display test
	maxY, maxX := screen.MaxYX()
	screen.MovePrint(maxY/2, maxX/2, "Hello, ncurses!")
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
