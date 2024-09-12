package tcellio

import (
	"fmt"

	"github.com/gdamore/tcell"
)

type TcellIO struct {
	Pixels [][]bool
	fg     int32
	bg     int32
	screen tcell.Screen
}

func (TcellIO) New(width, height byte, fgColor, bgColor int32) (*TcellIO, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return &TcellIO{}, fmt.Errorf("error in tcellio/Initialize(): %w", err)
	}

	pxs := make([][]bool, height)
	for i := range pxs {
		for j := byte(0); j <= width; j++ {
			pxs[i][j] = false
		}
	}
	tc := TcellIO{pxs, fgColor, bgColor, screen}

	return &tc, nil
}

func (io *TcellIO) SetPixel(int, int, bool) error {
	return nil
}

func (io *TcellIO) Refresh() error {
	return nil
}

func (io TcellIO) Listen() (byte, error) {
	return 0x0a, nil
}

func (io TcellIO) Terminate() error {
	return nil
}
