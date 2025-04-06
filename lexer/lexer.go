package lexer

import (
	"fmt"
	"unicode"
)

func EOF[T any](token Token[T]) bool {
	return token.Literal == "EOF"
}

type Token[T any] struct {
	Variant T
	Literal string
}

type Func[T any] func(lexer *Lexer[T], c uint8) (Token[T], bool, error)

type Params[T any] struct {
	Symbols map[uint8]T
	Ignore  ConditionFunc
}

func NewLexer[T any](params Params[T], delegates ...Func[T]) *Lexer[T] {
	return &Lexer[T]{
		ignored:   params.Ignore,
		symbols:   params.Symbols,
		delegates: delegates,
	}
}

type Lexer[T any] struct {
	ignored   ConditionFunc
	symbols   map[uint8]T
	delegates []Func[T]
	source    string
	cursor    int
}

func (l *Lexer[T]) Load(source string) {
	l.source = source
	l.cursor = 0
}

func (l *Lexer[T]) More() bool {
	return l.cursor < len(l.source)
}

func (l *Lexer[T]) Peek() (Token[T], error) {
	previous := l.cursor
	defer func() {
		l.cursor = previous
	}()
	return l.Next()
}

func (l *Lexer[T]) Next() (Token[T], error) {
	if l.cursor >= len(l.source) {
		return Token[T]{Literal: "EOF"}, nil
	}
	if err := l.Seek(Not(l.ignored)); err != nil {
		return Token[T]{}, err
	}
	if l.cursor >= len(l.source) {
		return Token[T]{Literal: "EOF"}, nil
	}
	character := l.source[l.cursor]
	if symbol, ok := l.symbols[character]; ok {
		l.cursor++
		return Token[T]{
			Variant: symbol,
			Literal: string(character),
		}, nil
	}
	for _, delegate := range l.delegates {
		token, ok, err := delegate(l, character)
		if err != nil {
			return Token[T]{}, err
		}
		if ok {
			return token, nil
		}
	}
	return Token[T]{}, fmt.Errorf("all delegates failed to process character '%s'", string(character))
}

func (l *Lexer[T]) literal(fn ConditionFunc) string {
	literal := ""
	for ; l.cursor < len(l.source) && fn(l.source[l.cursor]); l.cursor++ {
		literal += string(l.source[l.cursor])
	}
	return literal
}

func (l *Lexer[T]) Seek(c ConditionFunc) error {
	for ; l.cursor < len(l.source) && !c(l.source[l.cursor]); l.cursor++ {
	}
	if l.cursor == len(l.source) {
		return fmt.Errorf("seek could not find byte matching condition")
	}
	return nil
}

func Whitespace(c uint8) bool {
	return unicode.IsSpace(rune(c))
}

func Alphabetical(c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func Numerical(c uint8) bool {
	return c >= '0' && c <= '9'
}

func Alphanumeric(c uint8) bool {
	return Alphabetical(c) || Numerical(c)
}
