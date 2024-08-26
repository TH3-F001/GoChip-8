package greetings

import (
	"regexp"
	"testing"
)

func TestGreetName(t *testing.T) {
	name := "Gladys"
	want := regexp.MustCompile(`\b` + name + `\b`)
	msg, err := Greet("Gladys")
	if !want.MatchString(msg) || err != nil {
		t.Fatalf(`Greet("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	}
}

func TestGreetEmpty(t *testing.T) {
	msg, err := Greet("")
	if msg != "" || err == nil {
		t.Fatalf(`Hello("") = %q, %v, want "", error`, msg, err)
	}
}
