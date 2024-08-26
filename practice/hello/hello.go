package main

import (
	"fmt"
	"log"

	"example.com/greetings"
)

func main() {
	// Set properties of the logger including entry prefix and a flag to disable printing
	// include time, source file, and line number
	log.SetPrefix("greetings: ")
	log.SetFlags(0)

	message, err := greetings.RandomGreeting("F00L")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message)

	names := []string{
		"Yacko",
		"Wacko",
		"Dot",
		"Nurse",
	}

	greetings, err := greetings.GreetGroup(names)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, greeting := range greetings {
		fmt.Println(greeting)	
	}

}
