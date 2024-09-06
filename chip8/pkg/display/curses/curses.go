package curses

import (
	"log"

	"github.com/rthornton128/goncurses"
)

var screen goncurses.Window

func Initialize() {
	scr, err := goncurses.Init()
	if err != nil {
		log.Fatal("goncurses.init:", err)
	}
	screen = *scr
	screen.Clear()
	screen.Refresh()
	screen.MovePrint(5, 10, "Hello, ncurses!")
	screen.Refresh()

}

func Terminate() {
	goncurses.End()
}
