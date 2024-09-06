package ansi

import "fmt"

var display [32][64]string

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func Clear() {
	for i := 0; i < len(display); i++ {
		for j := 0; j < len(display[i]); j++ {
			display[i][j] = " "
			fmt.Print(display[i][j])
		}
		fmt.Print("\n")
	}
}

func Test() {
	moveCursor(10, 20)
	fmt.Println("Hello")
}

func Refresh() {
	for i := 0; i < len(display); i++ {
		for j := 0; j < len(display[i]); j++ {
			fmt.Print(display[i][j])
		}
		fmt.Print("\n")
	}
}
