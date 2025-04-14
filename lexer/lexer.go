package lexer

import (
	"fmt"
	"io"
	"unicode"
)

// EOF reports whether token represents the end of the source string or not.
func EOF[T any](token Token[T]) bool {
	return token.Literal == "EOF"
}

type Token[T any] struct {
	Variant T
	Literal string
}

type Func[T any] func(lexer *Lexer[T], c uint8) (Token[T], bool, error)

// Params represents the parameters required to construct a Lexer
type Params[T any] struct {
	// Symbols maps bytes to token variants such as '#': Pound.
	Symbols map[uint8]T
	// Ignore is a ConditionFunc that resolves to true for any byte that should be ignored provided that it is the first
	// byte considered for a token. Typically, this would include whitespace characters and linefeed.
	Ignore ConditionFunc[T]
}

// NewLexer constructs a Lexer that adhere to the provided Params and uses the delegates to map anything that is not
// already in the symbol table or being ignored.
func NewLexer[T any](params Params[T], delegates ...Func[T]) *Lexer[T] {
	return &Lexer[T]{
		ignored:   params.Ignore,
		symbols:   params.Symbols,
		delegates: delegates,
	}
}

// Lexer is a data structure that facilitates customizable lexical analysis based of functions passed into its
// constructor NewLexer.
type Lexer[T any] struct {
	ignored   ConditionFunc[T]
	symbols   map[uint8]T
	delegates []Func[T]
	source    string
	cursor    int
}

// Load configures the Lexer to read from the beginning of source when the next reading operation such as Next is
// invoked.
func (l *Lexer[T]) Load(source string) {
	l.source = source
	l.cursor = 0
}

// More reports whether there are more tokens to read. Do keep in mind that the remaining token could be the simply EOF.
func (l *Lexer[T]) More() bool {
	return l.cursor < len(l.source)
}

// Peek returns the result of calling Next but rewinds the internal cursor such that calling Peek or Next again will
// return the very same token.
func (l *Lexer[T]) Peek() (Token[T], error) {
	previous := l.cursor
	defer func() {
		l.cursor = previous
	}()
	return l.Next()
}

// Next returns the next available token and moves the cursor in preparation for subsequent invocations of Next.
func (l *Lexer[T]) Next() (Token[T], error) {
	if l.cursor >= len(l.source) {
		return Token[T]{}, io.EOF
	}
	if err := l.Seek(Not(l.ignored)); err != nil {
		return Token[T]{}, err
	}
	if l.cursor >= len(l.source) {
		return Token[T]{}, io.EOF
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

func (l *Lexer[T]) literal(fn ConditionFunc[T]) string {
	literal := ""
	for ; l.cursor < len(l.source) && fn(l, l.source[l.cursor]); l.cursor++ {
		literal += string(l.source[l.cursor])
	}
	return literal
}

// Seek places the internal cursor wherever the resulting byte at the cursor satisfies the ConditionFunc c.
func (l *Lexer[T]) Seek(c ConditionFunc[T]) error {
	for ; l.cursor < len(l.source) && !c(l, l.source[l.cursor]); l.cursor++ {
	}
	if l.cursor == len(l.source) {
		return io.EOF
	}
	return nil
}

// Whitespace is a ConditionFunc that returns true when c is a whitespace character.
func Whitespace[T any](lx *Lexer[T], c uint8) bool {
	return unicode.IsSpace(rune(c))
}

// Alphabetical is a ConditionFunc that returns true when c is alphabetical.
func Alphabetical[T any](lx *Lexer[T], c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// Numerical is a ConditionFunc that returns true when c is numerical.
func Numerical[T any](lx *Lexer[T], c uint8) bool {
	return c >= '0' && c <= '9'
}

// Alphanumeric is a ConditionFunc that returns true when c is alphabetical or numerical.
func Alphanumeric[T any](lx *Lexer[T], c uint8) bool {
	return Alphabetical(lx, c) || Numerical(lx, c)
}
