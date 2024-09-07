package display

type Display interface {
	Initialize() error
	Terminate() error
	SetPixel(int, int, bool) error
	Refresh() error
}
