package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

func isNewLine(b byte) bool { return b == '\n' || b == '\r' }
func isSpace(b byte) bool   { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }
func isLetter(b byte) bool  { return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' }
func isDigit(b byte) bool   { return '0' <= b && b <= '9' }
func isComment(b byte) bool { return b == '%' }
func isEscape(b byte) bool  { return b == '\\' }
func isGrouping(b byte) bool {
	return b == '{' || b == '}' || b == '[' || b == ']' || b == '(' || b == ')' || b == '<' || b == '>' || b == '$' || b == '@'
}

func consume(n int, data []byte) (advance int, token []byte, err error) {
	advance, token, err = n, data[:n], nil
	return
}

func consumeWhile(test func(byte) bool, data []byte) (advance int, token []byte, err error) {
	for i, b := range data {
		if !test(b) {
			advance, token, err = i, data[:i], nil
			break
		}
	}
	return
}

func consumeTill(test func(byte) bool, data []byte) (advance int, token []byte, err error) {
	for i, b := range data {
		if test(b) {
			advance, token, err = i, data[:i], nil
			break
		}
	}
	return
}

func splitter(data []byte, end bool) (advance int, token []byte, err error) {
	switch b := data[0]; {
	case isSpace(b):
		advance, token, err = consumeWhile(isSpace, data)
	case isLetter(b):
		advance, token, err = consumeWhile(isLetter, data)
	case isDigit(b):
		advance, token, err = consumeWhile(isDigit, data)
	case isGrouping(b):
		advance, token, err = consume(1, data)
	case isEscape(b):
		advance, token, err = consumeWhile(isLetter, data[1:])
		token = append([]byte("\\"), token...)
		advance++
	case isComment(b):
		advance, token, err = consumeTill(isNewLine, data)
	default:
		advance, token, err = consume(1, data)
	}
	return
}

// if bytes.HasPrefix(data[1:], []byte("start")) {
// 	advance, token, err = consumeLetters(data[1:])
// 	//FIXME advance++
// } else if strings.HasPrefix(data[1:], "begin") {
// 	advance, token, err = consumeLetters(data[5:])
// 	advance++
// } else if strings.HasPrefix(data[1:], "left") {
// 	advance, token, err = consumeCommand(data)
// } else {
// 	fallthrough
// }

func lex(s *bufio.Scanner) {
	s.Split(splitter)
	for s.Scan() {
		token := s.Text()
		fmt.Printf("%q\n", token)
	}

}

func main() {
	state := *new(State)
	fmt.Printf("%+v %T\n", state, state)

	reader := strings.NewReader("This is \ta \\LaTeX \\emph{test} string, containing newlines\nand some $math^42$. % It also includes a comment\nBy Tim~Steenvoorden.")
	scanner := bufio.NewScanner(reader)
	lex(scanner)

	for _, a := range os.Args[1:] {
		f, e := os.Open(a)
		if e != nil {
			fmt.Println(e)
		} else {
			s := bufio.NewScanner(f)
			lex(s)
		}
	}

}
