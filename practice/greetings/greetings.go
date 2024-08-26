package greetings

import (
	"errors"
	"fmt"

	"math/rand"
)

// Hello returns a greeting for a named person.

func Greet(name string) (string, error) {
	// return an error if no name is provided
	if name == "" {
		return "", errors.New("empty name not accepted")
	}

	// Return a greeting that inserts a name into a message
	message := fmt.Sprintf("Hello %v. Welcome!", name)
	return message, nil
}

func RandomGreeting(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name not accepted")
	}
	message := fmt.Sprintf(randomFormat(), name)
	return message, nil
}

// returns one of a set of greeting messages randomly
func randomFormat() string {
	formats := []string{
		"Hi %v. Welcome!",
		"Great to see you, %v!",
		"Hail, %v! Well met!",
	}

	return formats[rand.Intn(len(formats))]
}

// Returns a map that associates each named person with a greeting msg
func GreetGroup(names []string) (map[string]string, error) {
	// a map associates names with messages
	messages := make(map[string]string)

	// Loop through each in the slice calling the RandomGreet function for each
	for _, name := range names {
		if name == "Nurse" {
			message := fmt.Sprintf("Helloooooooooooo, %v! 😍", name)
			messages[name] = message
			continue
		}
		message, err := RandomGreeting(name)
		if err != nil {
			return nil, err
		}

		messages[name] = message
	}

	return messages, nil
}
