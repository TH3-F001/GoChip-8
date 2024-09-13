package sdlio

// NOTE: This package requires that libsdl2 is installed
import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

type SdlIO struct {
	Pixels [][]bool
	fg     uint32
	bg     uint32
	window *sdl.Window
}

func New(width, height byte, fgColor, bgColor uint32) (*SdlIO, error) {
	result := SdlIO{}
	if err := sdl.Init(uint32(sdl.INIT_EVERYTHING)); err != nil {
		return &result, err
	}

	win, err := sdl.CreateWindow("Keyboard Listener", int32(sdl.WINDOWPOS_UNDEFINED), int32(sdl.WINDOWPOS_UNDEFINED), 800, 600, uint32(sdl.WINDOW_SHOWN))
	if err != nil {
		return &result, err
	}

	result.window = win
	return &result, nil
}

func (io SdlIO) GetPixels() *[][]bool {
	return &io.Pixels
}

func (*SdlIO) SetPixel(row, col int, lit bool) error {
	return nil
}

func (*SdlIO) Refresh() error {
	return nil
}

func (*SdlIO) Listen() (byte, error) {
	active := true
	var result byte = 255
	var event sdl.Event
	for active {
		fmt.Println("doom")
		for event = sdl.WaitEvent(); event != nil; event = sdl.PollEvent() {
			fmt.Println(event)
			if keyboardEvent, ok := event.(*sdl.KeyboardEvent); ok {

				switch int(keyboardEvent.Keysym.Scancode) {

				case 0x82:
					result = 0x01
				case 0x83:
					result = 0x02
				case 0x84:
					result = 0x03
				case 0x85:
					result = 0x0C
				case 0x90:
					result = 0x04
				case 0x91:
					result = 0x05
				case 0x92:
					result = 0x06
				case 0x93:
					result = 0x0D
				case 0x9e:
					result = 0x07
				case 0x9f:
					result = 0x08
				case 0xa0:
					result = 0x09
				case 0xa1:
					result = 0x0E
				case 0xac:
					result = 0x0A
				case 0xad:
					result = 0x00
				case 0xae:
					result = 0x0B
				case 0xaf:
					result = 0x0F

				default:
					continue
				}
				if result != 255 {
					return result, nil
				}
			} else {
				fmt.Println("not okay")
			}
		}
	}
	return result, nil
}

func (io *SdlIO) Terminate() error {
	io.window.Destroy()
	sdl.Quit()
	return nil
}
