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
	set
	out
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

type lexer struct {
	src    string
	cursor int
}

func (l *lexer) peek() (token, error) {
	prev := l.cursor
	defer func() {
		l.cursor = prev
	}()
	return l.next()
}

func (l *lexer) next() (token, error) {
	l.literal(l.space)
	if l.cursor >= len(l.src) {
		return token{
			variant: eof,
		}, nil
	}
	char := l.src[l.cursor]
	if symbol, ok := symbols[char]; ok {
		l.cursor++
		return token{
			variant: symbol,
			literal: string(char),
		}, nil
	}
	switch char {
	case '/':
		if l.src[l.cursor+1] == '/' {
			if err := l.seek('\n'); err != nil {
				return token{}, err
			}
			return l.next()
		}
		return token{}, fmt.Errorf("invalid token '%s'", string(char))
	case '-':
		if l.src[l.cursor+1] == '>' {
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
func (l *lexer) seek(c uint8) error {
	for ; l.cursor < len(l.src) && l.src[l.cursor] != c; l.cursor++ {
	}
	if l.cursor == len(l.src) {
		return fmt.Errorf("character '%s' not found", string(c))
	}
	return nil
}

func (l *lexer) space(c uint8) bool {
	return unicode.IsSpace(rune(c))
}

func (l *lexer) literal(fn func(uint8) bool) string {
	literal := ""
	for ; l.cursor < len(l.src) && fn(l.src[l.cursor]); l.cursor++ {
		literal += string(l.src[l.cursor])
	}
	return literal
}

func (l *lexer) identifier(c uint8) bool {
	if unicode.IsSpace(rune(c)) {
		return false
	}
	if alphanumerical(c) {
		return true
	}
	switch c {
	case '_', '.':
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
