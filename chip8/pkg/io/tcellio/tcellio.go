package tcellio

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

// TcellIO ... Holds state for the active tcellio display. Sh-ould be instantiated using tcellio.New()
// pixels is a 2-dimensional array of bools representing whether a pixel should be on or off. It's size should be stack once initialized
// fg and bg are hex values for the display color
// screen holds the active tcell.Screen instance
// style holds the tcell.Style instance
// maxRow and maxCol hold the largest column/row that can be written to. used to limit excessive use of len(pixels) - 1
type TcellIO struct {
	pixels [][]bool
	fg     uint32
	bg     uint32
	screen tcell.Screen
	style  tcell.Style
	maxRow int
	maxCol int
}

// charMap... A simple booleon map of true/false to on/off pixel runes. Used to prevent excessive if statements
// Suggested pixel characters: ░▒▓
var charMap map[bool]rune = map[bool]rune{
	true:  '░',
	false: ' ',
}

// New ... Creates a new tcell screen instance, initializes color and screen size, and returns a TCellio instance
// fg and bg expects colors as hexcodes. red green and blue are split from the hex, and a new tcell color is created
// returns an error if: rows or cols are less or equal to zero, or if screen creation/initialization fails
func New(rows, cols int, fgColor, bgColor uint32) (*TcellIO, error) {
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("error in tcellio/New(): Width/Height must be more than zero: rows=%v, cols=%v", rows, cols)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("error in tcellio/New(): %w", err)
	}

	if err := screen.Init(); err != nil {
		screen.Fini()
		return nil, fmt.Errorf("error in tcellio/New(): %w", err)
	}

	pxs := make([][]bool, rows)
	for row := range pxs {
		pxs[row] = make([]bool, cols)
		for col := range pxs[row] {
			pxs[row][col] = false
		}
	}

	fg := tcell.NewHexColor(int32(fgColor))
	bg := tcell.NewHexColor(int32(bgColor))

	style := tcell.StyleDefault.Background(bg).Foreground(fg)
	screen.SetStyle(style)

	tc := TcellIO{
		pixels: pxs,
		fg:     fgColor,
		bg:     bgColor,
		screen: screen,
		style:  style,
		maxRow: rows - 1,
		maxCol: cols - 1,
	}

	return &tc, nil
}

// GetMaxRow ... Returns the highest pixels row that can be written to (if len(pixels[row] is 32, then maxCol is 31).
func (io TcellIO) GetMaxRow() int {
	return io.maxRow
}

// GetMaxCol ... Returns the highest pixels column that can be written to (if len(pixels[row] is 32, then maxCol is 31).
func (io TcellIO) GetMaxCol() int {
	return io.maxCol
}

// GetPixels ... Returns a copy of the TcellIO.pixels 'array' (slice)
func (io TcellIO) GetPixels() [][]bool {
	return io.pixels
}

// GetPixel ... Returns true if the pixel at the given cell in the pixels array is on, false if it's off, and an error if out of bounds
func (io TcellIO) GetPixel(row, col int) (bool, error) {
	if row < 0 || row > io.maxRow || col < 0 || col > io.maxCol {
		return false, fmt.Errorf("out of bounds request to TcellIO.GetPixel() . request to get pixel at (row:%d, col:%d) is out of bounds. max coordinates: (row:%d, col:%d)", row, col, io.maxRow, io.maxCol)
	}
	return io.pixels[row][col], nil
}

// SetPixel ... Sets the pixel at a given row and column to either on or off. returns an error if out of bounds
func (io *TcellIO) SetPixel(row, col int, lit bool) error {
	if row > io.maxRow || row < 0 || col > io.maxCol || col < 0 {
		return fmt.Errorf("overflow detected in TcellIO.SetPixel(). request to set (row:%d, col:%d) to %v is out of bounds. max coordinates: (row:%d, col:%d)", row, col, lit, io.maxRow, io.maxCol)
	}
	io.pixels[row][col] = lit

	return nil
}

// Refresh ...Iterates over the display's array of pixels, and updates the display to match the array. error is only to conform to the IO interface and isnt used
func (io *TcellIO) Refresh() error {
	for row := range io.pixels {
		for col := range io.pixels[row] {
			io.screen.SetContent(col, row, charMap[io.pixels[row][col]], nil, io.style)
		}
	}
	io.screen.Show()
	return nil // Satisfies the interface
}

// Listen ... listens for the traditonal Chip8 keypresses eturns the corresponding hex code and forwards any errors.
func (io TcellIO) Listen() (byte, error) {
	event := io.screen.PollEvent()
	switch event := event.(type) {
	case *tcell.EventResize:
		io.screen.Sync()
	case *tcell.EventKey:
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

// ListenForTermination ... Specifically listens for user interupts to terminate the program. (meant to be run concurrently)
func (io TcellIO) ListenForTermination(termCh chan<- bool) {
	go func() {
		for {
			event := io.screen.PollEvent()
			switch event := event.(type) {
			case *tcell.EventKey:
				if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
					termCh <- true
					return // Exit the goroutine when termination key is pressed
				}
			}
		}
	}()
}

// Terminate ... Clears the screen, destroys it, and exits the program
func (io *TcellIO) Terminate() {
	io.screen.Clear()
	io.screen.Fini()
	os.Exit(1)
}
