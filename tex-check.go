package main

import "fmt"

type (
	Symbol interface {
		opening() string
		closing() string
	}
	Brace     struct{}
	Bracket   struct{}
	Paren     struct{}
	Chevron   struct{}
	Dollar    struct{}
	At        struct{}
	Delimiter struct{}
	StartStop string
	BeginEnd  string
)

func (_ Brace) opening() string     { return "{" }
func (_ Bracket) opening() string   { return "[" }
func (_ Paren) opening() string     { return "(" }
func (_ Chevron) opening() string   { return "<" }
func (_ Dollar) opening() string    { return "$" }
func (_ At) opening() string        { return "@" }
func (_ Delimiter) opening() string { return "\\left" }
func (s StartStop) opening() string { return "\\start" + string(s) }
func (s BeginEnd) opening() string  { return "\\begin{" + string(s) + "}" }

func (_ Brace) closing() string     { return "}" }
func (_ Bracket) closing() string   { return "]" }
func (_ Paren) closing() string     { return ")" }
func (_ Chevron) closing() string   { return ">" }
func (_ Dollar) closing() string    { return "$" }
func (_ At) closing() string        { return "@" }
func (_ Delimiter) closing() string { return "\\right" }
func (s StartStop) closing() string { return "\\stop" + string(s) }
func (s BeginEnd) closing() string  { return "\\end{" + string(s) + "}" }

type (
	Mode  uint
	Line  uint
	Stack []Symbol
	State struct {
		mode  Mode
		line  Line
		stack Stack
	}
)

const (
	Normal Mode = iota
	Math
)

func main() {
	state := *new(State)
	fmt.Printf("%+v %T\n", state, state)
}
