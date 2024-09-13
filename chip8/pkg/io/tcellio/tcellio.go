package tcellio

import (
	"fmt"

	"github.com/gdamore/tcell"
)

type TcellIO struct {
	Pixels [][]bool
	fg     uint32
	bg     uint32
	screen tcell.Screen
	style  tcell.Style
}

// ░▒▓
const on rune = '▒'
const off rune = ' '

// tcellio.New() creates a new tcell instance, initializes color and screen size, and returns a TCellio instance
// fg and bg expects colors as hexcodes. red green and blue are split from the hex, and a new tcell color is created
func New(width, height byte, fgColor, bgColor uint32) (*TcellIO, error) {
	// Create new TCell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return &TcellIO{}, fmt.Errorf("error in tcellio/New(): %w", err)
	}

	// Initialize the screen
	if err := screen.Init(); err != nil {
		return &TcellIO{}, fmt.Errorf("error in tcellio/New(): %w", err)
	}

	// Build Pixels array
	pxs := make([][]bool, height)
	for row := range pxs {
		pxs[row] = make([]bool, width)
		for col := byte(0); col < width; col++ {
			pxs[row][col] = false
		}
	}

	// Set up colors

	fg := tcell.NewHexColor(int32(fgColor))
	bg := tcell.NewHexColor(int32(bgColor))

	style := tcell.StyleDefault.Background(bg).Foreground(fg)
	screen.SetStyle(style)

	tc := TcellIO{pxs, fgColor, bgColor, screen, style}

	return &tc, nil
}

func (io TcellIO) GetPixels() *[][]bool {
	return &io.Pixels
}

func (io *TcellIO) SetPixel(row, col int, lit bool) error {
	io.Pixels[row][col] = lit
	return nil
}

func (io *TcellIO) Refresh() error {
	for row := range io.Pixels {
		for col := range io.Pixels[row] {
			if !io.Pixels[row][col] {
				io.screen.SetContent(col, row, off, nil, io.style)
			} else {
				io.screen.SetContent(col, row, on, nil, io.style)
			}
		}
	}
	io.screen.Show()
	return nil
}

func (io TcellIO) Listen() (byte, error) {
	event := io.screen.PollEvent()
	switch event := event.(type) {
	case *tcell.EventResize:
		io.screen.Sync()
	case *tcell.EventKey:
		//
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
			io.Terminate()
			return 255, fmt.Errorf("exited due to user keyboard interupt")
		}
		switch event.Rune() {
		case '1':
			return 0x1, nil
		case '2':
			return 0x2, nil
		case '3':
			return 0x3, nil
		case '4':
			return 0xC, nil
		case 'q':
			return 0x4, nil
		case 'w':
			return 0x5, nil
		case 'e':
			return 0x6, nil
		case 'r':
			return 0xD, nil
		case 'a':
			return 0x7, nil
		case 's':
			return 0x8, nil
		case 'd':
			return 0x9, nil
		case 'f':
			return 0xE, nil
		case 'z':
			return 0xA, nil
		case 'x':
			return 0x0, nil
		case 'c':
			return 0xB, nil
		case 'v':
			return 0xF, nil

		}
	}
	return 0x0a, nil
}

func (io *TcellIO) Terminate() error {
	io.screen.Fini()
	return nil
}
