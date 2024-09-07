package ansi

import (
	"errors"
	"fmt"
)

const on rune = '█'
const off rune = ' '

type ADisplay struct {
	pixels [][]rune
	fg     int16
	bg     int16
}

func (d *ADisplay) Initialize() error {
	var err error
	for i := 0; i < len(d.pixels); i++ {
		for j := 0; j < len(d.pixels[i]); j++ {
			d.pixels[i][j] = off
			_, err = fmt.Print(d.pixels[i][j])
		}
		_, err = fmt.Print("\n")
	}
	return err
}

func (d *ADisplay) SetPixel(row, col int, lit bool) error {
	if row < 0 || row >= len(d.pixels) || col < 0 || col >= len(d.pixels[row]) {
		return errors.New("Error in ansi/ADisplay.SetPixel()\tDisplay coordinates out of bounds")
	}

	px := &d.pixels[row][col]
	if lit {
		*px = on
	} else {
		*px = off
	}
	return nil
}

func (d *ADisplay) Refresh() {
	for row := 0; row < len(d.pixels); row++ {
		for col := 0; col < len(d.pixels[row]); col++ {
			moveCursor(row, col)
			fmt.Print(d.pixels[row][col])
		}
		fmt.Print("\n")
	}
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}
