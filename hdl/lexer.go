package hdl

import (
	"fmt"
	"unicode"
)

var (
	symbols = map[uint8]variant{
		'(': leftParenthesis,
		')': rightParenthesis,
		'=': equals,
		':': colon,
		',': comma,
		'{': leftCurlyBrace,
		'}': rightCurlyBrace,
		'[': leftBracket,
		']': rightBracket,
		'.': dot,
	}
	keywords = map[string]variant{
		"chip": chip,
		"set":  set,
		"out":  out,
	}
)

const (
	eof variant = iota
	chip
	out
	set
	dot
	identifier
	integer
	leftParenthesis
	rightParenthesis
	colon
	comma
	arrow
	leftCurlyBrace
	rightCurlyBrace
	leftBracket
	rightBracket
	equals
)

type variant int

type token struct {
	variant variant
	literal string
}

type Lexer struct {
	Source string
	cursor int
}

func (l *Lexer) peek() (token, error) {
	prev := l.cursor
	defer func() {
		l.cursor = prev
	}()
	return l.next()
}

func (l *Lexer) next() (token, error) {
	l.literal(l.space)
	if l.cursor >= len(l.Source) {
		return token{
			variant: eof,
		}, nil
	}
	char := l.Source[l.cursor]
	if symbol, ok := symbols[char]; ok {
		l.cursor++
		return token{
			variant: symbol,
			literal: string(char),
		}, nil
	}
	switch char {
	case '/':
		if l.Source[l.cursor+1] == '/' {
			if err := l.seek('\n'); err != nil {
				return token{}, err
			}
			return l.next()
		}
		return token{}, fmt.Errorf("invalid token '%s'", string(char))
	case '-':
		if l.Source[l.cursor+1] == '>' {
			l.cursor += 2
			return token{
				variant: arrow,
				literal: "->",
			}, nil
		}
		return token{}, fmt.Errorf("invalid token '%s'", string(char))
	}
	if !alphanumerical(char) {
		return token{}, fmt.Errorf("invalid token '%s'", string(char))
	}
	// Identifiers cannot start with a digit, therefore we must first check if the current character is an integer to
	// decide whether to regard this as an integer Literal
	if numerical(char) {
		return token{
			variant: integer,
			literal: l.literal(numerical),
		}, nil
	}
	literal := l.literal(l.identifier)
	if keyword, ok := keywords[literal]; ok {
		return token{
			variant: keyword,
			literal: literal,
		}, nil
	}
	return token{
		variant: identifier,
		literal: literal,
	}, nil
}

// seek places the cursor at the next instance of the supplied character, skipping anything before finding a match
func (l *Lexer) seek(c uint8) error {
	for ; l.cursor < len(l.Source) && l.Source[l.cursor] != c; l.cursor++ {
	}
	if l.cursor == len(l.Source) {
		return fmt.Errorf("character '%s' not found", string(c))
	}
	return nil
}

func (l *Lexer) space(c uint8) bool {
	return unicode.IsSpace(rune(c))
}

func (l *Lexer) literal(fn func(uint8) bool) string {
	literal := ""
	for ; l.cursor < len(l.Source) && fn(l.Source[l.cursor]); l.cursor++ {
		literal += string(l.Source[l.cursor])
	}
	return literal
}

func (l *Lexer) identifier(c uint8) bool {
	if unicode.IsSpace(rune(c)) {
		return false
	}
	if alphanumerical(c) {
		return true
	}
	switch c {
	case '_':
		return true
	default:
		return false
	}
}

func alphanumerical(c uint8) bool {
	return alphabetical(c) || numerical(c)
}

func alphabetical(c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func numerical(c uint8) bool {
	return c >= '0' && c <= '9'
}
