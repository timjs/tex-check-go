package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
	Other     byte
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
func (b Other) opening() string     { return string(b) }
func (s StartStop) opening() string { return "\\start" + string(s) }
func (s BeginEnd) opening() string  { return "\\begin{" + string(s) + "}" }

func (_ Brace) closing() string     { return "}" }
func (_ Bracket) closing() string   { return "]" }
func (_ Paren) closing() string     { return ")" }
func (_ Chevron) closing() string   { return ">" }
func (_ Dollar) closing() string    { return "$" }
func (_ At) closing() string        { return "@" }
func (_ Delimiter) closing() string { return "\\right" }
func (b Other) closing() string     { return string(b) }
func (s StartStop) closing() string { return "\\stop" + string(s) }
func (s BeginEnd) closing() string  { return "\\end{" + string(s) + "}" }

type (
	Mode  uint
	Line  uint
	Stack []LocatedSymbol
	State struct {
		mode  Mode
		line  Line
		stack Stack
	}
	LocatedSymbol struct {
		symbol Symbol
		line   Line
	}
)

const (
	NORMAL Mode = iota
	MATH
	VERBATIM
)

func isNewLine(b byte) bool { return b == '\n' || b == '\r' }
func isSpace(b byte) bool   { return b == ' ' || b == '\t' || b == '\v' || b == '\f' }
func isLetter(b byte) bool  { return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' }
func isDigit(b byte) bool   { return '0' <= b && b <= '9' }
func isEscape(b byte) bool  { return b == '\\' }
func isComment(b byte) bool { return b == '%' }
func isGrouping(b byte) bool {
	return b == '{' || b == '}' || b == '[' || b == ']' || b == '(' || b == ')' || b == '<' || b == '>' || b == '$' || b == '@'
}

func consume(n int, data []byte) (advance int, token []byte, err error) {
	advance, token, err = n, data[:n], nil
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

func consumeWhile(test func(byte) bool, data []byte) (advance int, token []byte, err error) {
	for i, b := range data {
		if !test(b) {
			advance, token, err = i, data[:i], nil
			break
		}
	}
	return
}

func symbolise(b byte) Symbol {
	switch b {
	case '{', '}':
		return Brace{}
	case '[', ']':
		return Bracket{}
	case '(', ')':
		return Paren{}
	case '<', '>':
		return Chevron{}
	default:
		return Other(b)
	}
}

func splitter(data []byte, end bool) (advance int, token []byte, err error) {
	switch b := data[0]; {
	case isNewLine(b):
		advance, token, err = consume(1, data)
	case isSpace(b):
		advance, token, err = consumeWhile(isSpace, data)
	case isLetter(b):
		advance, token, err = consumeWhile(isLetter, data)
	case isDigit(b):
		advance, token, err = consumeWhile(isDigit, data)
	case isEscape(b):
		advance, token, err = consumeWhile(isLetter, data[1:])
		token = append([]byte{'\\'}, token...)
		advance++
	case isComment(b):
		advance, token, err = consumeTill(isNewLine, data)
	default:
		advance, token, err = consume(1, data)
	}
	return
}

func balanced(scanner *bufio.Scanner) bool {
	// state := State{mode: NORMAL, line:  1, stack: *new(Stack)}
	state := new(State)
	state.line++
	scanner.Split(splitter)

	for scanner.Scan() {
		switch state.mode {
		case NORMAL, MATH:
			switch token := scanner.Bytes(); token[0] {
			case '\n', '\r':
				state.line++
			case '\\':
				switch {
				case bytes.Equal(token, []byte("\\starttyping")):
					state.push(StartStop("typing"))
					state.mode = VERBATIM
				case bytes.Equal(token, []byte("\\type")):
					scanner.Scan() // delimiter
					state.push(symbolise(scanner.Bytes()[0]))
					state.mode = VERBATIM
				case bytes.HasPrefix(token, []byte("\\start")):
					name := bytes.TrimPrefix(token, []byte("\\start"))
					state.push(StartStop(name))
				case bytes.HasPrefix(token, []byte("\\stop")):
					name := bytes.TrimPrefix(token, []byte("\\stop"))
					state.pop(StartStop(name))
				case bytes.Equal(token, []byte("\\begin")):
					scanner.Scan() // '{'
					scanner.Scan() // name
					state.push(BeginEnd(scanner.Bytes()))
					scanner.Scan() // '{'
				case bytes.Equal(token, []byte("\\end")):
					scanner.Scan() // '{'
					scanner.Scan() // name
					state.pop(BeginEnd(scanner.Bytes()))
					scanner.Scan() // '{'
				case bytes.Equal(token, []byte("\\left")):
					scanner.Scan() // delimiter
					state.push(Delimiter{})
				case bytes.Equal(token, []byte("\\right")):
					scanner.Scan() // delimiter
					state.pop(Delimiter{})
				}
			case '{':
				state.push(Brace{})
			case '}':
				state.pop(Brace{})
			case '[':
				state.push(Bracket{})
			case ']':
				state.pop(Bracket{})
			case '(':
				state.push(Paren{})
			case ')':
				state.pop(Paren{})
			case '$':
				switch state.mode {
				case MATH:
					state.pop(Dollar{})
					state.mode = NORMAL
				case NORMAL:
					state.push(Dollar{})
					state.mode = MATH
				}
			case '@':
				state.push(At{})
				state.mode = VERBATIM
			}
		case VERBATIM:
			last := state.peak()
			if scanner.Text() == last.symbol.closing() {
				state.mode = NORMAL
				state.pop(last.symbol)
			}
		}
	}
	if len(state.stack) != 0 {
		last := state.peak()
		fmt.Printf("!! Unexpected end of file, expected %q\n   (to close %q from line %d)\n",
			last.symbol.closing(), last.symbol.opening(), last.line)
		return false
	} else {
		return true
	}
}

func (state *State) push(symbol Symbol) {
	// fmt.Printf("++ %v %T\n", symbol, symbol)
	state.stack = append(state.stack, LocatedSymbol{symbol, state.line})
}

func (state *State) pop(symbol Symbol) {
	// fmt.Printf("-- %v %T\n", symbol, symbol)
	if len(state.stack) == 0 {
		fmt.Printf("!! Line %d:\n   Unexpected %q, closed without opening\n",
			state.line, symbol.closing())
	} else {
		last := state.peak()
		if symbol == last.symbol {
			state.stack = state.stack[:len(state.stack)-1]
		} else {
			fmt.Printf("!! Line %d:\n   Unexpected %q, expected %q\n   (to close %q from line %d)\n",
				state.line, symbol.closing(), last.symbol.closing(), last.symbol.opening(), last.line)
		}
	}
}

func (state *State) peak() LocatedSymbol {
	return state.stack[len(state.stack)-1]
}

func lex(s *bufio.Scanner) {
	s.Split(splitter)
	for s.Scan() {
		token := s.Text()
		fmt.Printf("%q\n", token)
	}
}

func main() {

	// reader := strings.NewReader("This is \ta \\LaTeX \\emph{test} string, containing newlines\nand some $math^42$. % It also includes a comment\nBy Tim~Steenvoorden.")
	// scanner := bufio.NewScanner(reader)
	// balanced(scanner)

	for _, a := range os.Args[1:] {
		f, e := os.Open(a)
		if e != nil {
			fmt.Println(e)
		} else {
			s := bufio.NewScanner(f)
			// lex(s)
			fmt.Printf(">> %s...\n", a)
			balanced(s)
		}
	}

}
