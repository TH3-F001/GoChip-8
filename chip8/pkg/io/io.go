package io

// Why combine display and input into one? because most libraries typically handle bot in a semi coupled manner.
// SDL requires a window to take scan codes, and TCell handles keyboard events itself, not externally
// at the end of the day, code doesnt always reflect reality.

type IO interface {
	Initialize(int, int, int32, int32) (IO, error)
	SetPixel(int, int, bool) error
	Listen() byte
	Terminate() error
}
