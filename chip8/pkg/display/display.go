package display

type Display interface {
	NewDisplay(int, int, int32, int32) (Display, error)
	Terminate() error
	SetPixel(int, int, bool) error
	Refresh() error
}
