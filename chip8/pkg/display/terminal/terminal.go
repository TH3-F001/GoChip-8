package terminal

import "fmt"

var display [64][32]int

func test_display() {
	for _, row := range display {
		fmt.Println(row)
	}
}
