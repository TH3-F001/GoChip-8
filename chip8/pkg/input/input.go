package input

type Input interface {
	Initialize() error
	Listen() byte
	Terminate() error
}
