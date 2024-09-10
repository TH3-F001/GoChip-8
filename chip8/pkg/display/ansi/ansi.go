package ansi

import (
	"errors"
	"fmt"
)

const on rune = '█'
const off rune = ' '

type aDisplay struct {
	Pixels [][]rune
	fg     int32
	bg     int32
}

// Creates and returns a new aDisplay struct given max width, height, and colors
func NewDisplay(width, height int, fgColor, bgColor int32) (aDisplay, error) { // ✅
	if width <= 0 || height <= 0 {
		return aDisplay{}, fmt.Errorf("error in ansi/aDisplay.NewDisplay(): display must be at least 1px wide and 1px tall. supplied size: %dx%d", width, height)
	}

	if fgColor < 0 || fgColor >= 255 || bgColor < 0 || bgColor >= 255 {
		return aDisplay{}, fmt.Errorf("error in ansi/aDisplay.NewDisplay(): aDisplay only supports colors between 0 and 255. supplied colors: fg=%d, bg=%d", fgColor, bgColor)
	}

	pxs := make([][]rune, height)
	for i := range pxs {
		for j := 0; j <= width; j++ {
			pxs[i][j] = off
		}
	}

	// hide cursor while the display is active, make it visible on termination
	fmt.Print("\033[?25l")

	return aDisplay{pxs, fgColor, bgColor}, nil
}

// Sets a pixel in the aDisplay.Pixels array to on or off given row and column coordinates
func (d *aDisplay) SetPixel(row, col int, lit bool) error { // ✅
	if row < 0 || row >= len(d.Pixels) || col < 0 || col >= len(d.Pixels[row]) {
		return errors.New("error in ansi/aDisplay.setpixel(): display coordinates out of bounds")
	}

	px := &d.Pixels[row][col]
	if lit {
		*px = on
	} else {
		*px = off
	}
	return nil
}

// Refreshes the display to reflect the contents of display.Pixels (does notpackage ansi
func (d aDisplay) Refresh() error {
	if len(d.Pixels) <= 0 || len(d.Pixels[0]) <= 0 {
		return fmt.Errorf("error in ansi/aDisplay.Refresh(): display size is not in displayable bounds. display size: %dx%d, expected values over zero.", len(d.Pixels[0]), len(d.Pixels))
	}
	for row := 0; row < len(d.Pixels); row++ {
		for col := 0; col < len(d.Pixels[row]); col++ {
			moveCursor(row, col)
			fmt.Print(d.Pixels[row][col])
		}
		fmt.Print("\n")
	}
	return nil
}

func (d aDisplay) Terminate() error {
	// make the cursor visible
	fmt.Print("\033[?25h")
	return nil
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}
