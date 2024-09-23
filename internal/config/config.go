package config

type Config struct {
	IOType                string
	DefaultFont           string
	FgColor               uint32
	BgColor               uint32
	InstructionsPerSecond uint32
	CosmacCompatible      bool
	VerticalWrapping      bool
	ProgramPath           string
}
