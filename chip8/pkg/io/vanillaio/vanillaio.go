package vanillaio

import (
	"errors"
	"fmt"
)

type VanillaIO struct {
	Pixels [][]bool
	fg     uint32
	bg     uint32
}

func New(width, height byte, fgColor, bgColor uint32) (*VanillaIO, error) {
	if width <= 0 || height <= 0 {
		return &VanillaIO{}, fmt.Errorf("error in ansi/VanillaIO.NewDisplay(): display must be at least 1px wide and 1px tall. supplied size: %dx%d", width, height)
	}

	if fgColor < 0 || fgColor >= 255 || bgColor < 0 || bgColor >= 255 {
		return &VanillaIO{}, fmt.Errorf("error in ansi/VanillaIO.NewDisplay(): VanillaIO only supports colors between 0 and 255. supplied colors: fg=%d, bg=%d", fgColor, bgColor)
	}

	pxs := make([][]bool, height)
	for i := range pxs {
		for j := byte(0); j <= width; j++ {
			pxs[i][j] = false
		}
	}

	// hide cursor while the display is active, make it visible on termination
	fmt.Print("\033[?25l")

	return &VanillaIO{pxs, fgColor, bgColor}, nil
}

func (io VanillaIO) GetPixels() *[][]bool {
	return &io.Pixels
}

func (io *VanillaIO) SetPixel(row, col int, lit bool) error {
	if row < 0 || row >= len(io.Pixels) || col < 0 || col >= len(io.Pixels[row]) {
		return errors.New("error in ansi/VanillaIO.setpixel(): display coordinates out of bounds")
	}

	px := &io.Pixels[row][col]
	if lit {
		*px = true
	} else {
		*px = false
	}
	return nil
}

func (io VanillaIO) Listen() (byte, error) {
	return 0x0a, nil
}

func (io VanillaIO) Terminate() error {
	// make the cursor visible
	fmt.Print("\033[?25h")
	return nil
}

// Refreshes the display to reflect the contents of display.Pixels (does notpackage ansi
func (d VanillaIO) Refresh() error {
	if len(d.Pixels) <= 0 || len(d.Pixels[0]) <= 0 {
		return fmt.Errorf("error in ansi/VanillaIO.Refresh(): display size is not in displayable bounds. display size: %dx%d, expected values over zero.", len(d.Pixels[0]), len(d.Pixels))
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

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}
