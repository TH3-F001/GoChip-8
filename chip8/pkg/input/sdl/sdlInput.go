package sdlInput

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

type SdlInput struct{}

func Initialize() error {
	if err := sdl.Init(uint32(sdl.INIT_EVENTS)); err != nil {
		return err
	}
	return nil
}

func Listen() byte {
	active := true
	var result byte = 255
	var event sdl.Event
	for active {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if keyboardEvent, ok := event.(*sdl.KeyboardEvent); ok {
				fmt.Println(keyboardEvent.Keysym.Scancode)
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
					active = false
				}
			}
		}
	}
	return result
}

func Terminate() error {
	sdl.Quit()
	return nil
}
