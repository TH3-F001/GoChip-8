package ansi

import (
	"errors"
	"fmt"
)

const on rune = '█'
const off rune = ' '

type aDisplay struct {
	pixels [][]rune
	fg     int32
	bg     int32
}

// Creates and returns a new aDisplay struct given max width, height, and colors
func NewDisplay(width, height int, fgColor, bgColor int32) (aDisplay, error) { // ✅
	if width <= 0 || height <= 0 {
		return aDisplay{}, fmt.Errorf("error in ansi/aDisplay.NewDisplay(): display must be at least 1px wide and 1px tall. supplied size: %dx%d", width, height)
	}

	if fgColor < 0 || fgColor >= 16 || bgColor < 0 || bgColor >= 16 {
		return aDisplay{}, fmt.Errorf("error in ansi/aDisplay.NewDisplay(): aDisplay only supports colors between 0 and 15. supplied colors: fg=%d, bg=%d", fgColor, bgColor)
	}

	pxs := make([][]rune, height)
	for i := range pxs {
		for j := 0; j <= width; j++ {
			pxs[i][j] = off
		}
	}

	return aDisplay{pxs, fgColor, bgColor}, nil
}

// Sets a pixel in the aDisplay.pixels array to on or off given row and column coordinates
func (d *aDisplay) SetPixel(row, col int, lit bool) error { // ✅
	if row < 0 || row >= len(d.pixels) || col < 0 || col >= len(d.pixels[row]) {
		return errors.New("error in ansi/aDisplay.setpixel(): display coordinates out of bounds")
	}

	px := &d.pixels[row][col]
	if lit {
		*px = on
	} else {
		*px = off
	}
	return nil
}

// Refreshes the display to reflect the contents of display.pixels (does notpackage ansi
func (d aDisplay) Refresh() error {
	if len(d.pixels) <= 0 || len(d.pixels[0]) <= 0 {
		return fmt.Errorf("error in ansi/aDisplay.Refresh(): display size is not in displayable bounds. display size: %dx%d, expected values over zero.", len(d.pixels[0]), len(d.pixels))
	}
	for row := 0; row < len(d.pixels); row++ {
		for col := 0; col < len(d.pixels[row]); col++ {
			moveCursor(row, col)
			fmt.Print(d.pixels[row][col])
		}
		fmt.Print("\n")
	}
	return nil
}

func (d aDisplay) Terminate() error {
	return nil
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}
