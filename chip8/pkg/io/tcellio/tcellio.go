package tcellio

import "github.com/gdamore/tcell"

type Tcellio struct {
	Pixels [][]rune
	fg     int32
	bg     int32
}

func (io *Tcellio) Initialize(int, int, int32, int32) (Tcellio, error) {
	return Tcellio{}, nil
}

func (io *Tcellio) SetPixel(int, int, bool) error {
	return nil
}

func (io Tcellio) Listen() (byte, error) {
	return 0x0a, nil
}

func (io Tcellio) Terminate() error {
	return nil
}




