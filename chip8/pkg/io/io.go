package io

// Why combine display and input into one? because most libraries typically handle bot in a semi coupled manner.
// SDL requires a window to take scan codes, and TCell handles keyboard events itself, not externally
// at the end of the day, code doesnt always reflect reality.

// IO ... Handles User Input, Video Output, and Sound output for the Chip-8
type IO interface {

	// GetMaxRow ... Returns the highest row that can be accessed on the display's array of pixels
	GetMaxRow() int

	// GetMaxCol ... Returns the highest column that can be accessed on the display's array of pixels
	GetMaxCol() int

	// GetPixels ... returns a copy of the display's pixel array
	GetPixels() [][]bool

	// GetPixel ... Returns true if the pixel at the given cell in the pixels array is on, false if it's off, and an error if out of bounds
	GetPixel(row, col int) (bool, error)

	// SetPixel ... Sets the pixel at a given row and column to either on or off. returns an error if out of bounds
	SetPixel(row, col int, lit bool) error

	// Refresh ... Iterates over the display's array of pixels, and updates the display to match the array. and forwards any errors
	Refresh() error

	// Listen ... listens for the traditonal Chip8 keypresses eturns the corresponding hex code and forwards any errors. (run concurrently for no blockers, else run vanilla)
	Listen() (byte, error)

	// ListenForTermination ... Specifically listens for user interupts to terminate the program. (meant to be run concurrently)
	ListenForTermination(termCh chan<- bool)

	// Terminate ... Clears the screen, destroys it, and exits the program
	Terminate()
}
