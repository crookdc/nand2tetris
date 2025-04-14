package vm

import (
	"github.com/crookdc/nand2tetris/lexer"
	"io"
)

const (
	push variant = iota
	pop
	add
	sub
	neg
	eq
	gt
	lt
	and
	or
	not
	constant
	static
	pointer
	local
	arg
	tmp
	this
	that
	integer
	fn
	identifier
	ret
)

func newLexer(r io.Reader) (*lexer.Lexer[variant], error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	l := lexer.NewLexer[variant](
		lexer.Params[variant]{
			Symbols: make(map[uint8]variant),
			Ignore: lexer.Any(
				lexer.Whitespace[variant],
				lexer.LineComment[variant]("//"),
			),
		},
		lexer.Integer[variant](integer),
		lexer.Keywords[variant](
			map[string]variant{
				"push":     push,
				"pop":      pop,
				"add":      add,
				"sub":      sub,
				"neg":      neg,
				"and":      and,
				"or":       or,
				"not":      not,
				"eq":       eq,
				"lt":       lt,
				"gt":       gt,
				"constant": constant,
				"local":    local,
				"static":   static,
				"argument": arg,
				"temp":     tmp,
				"this":     this,
				"that":     that,
				"pointer":  pointer,
				"function": fn,
				"return":   ret,
			},
			lexer.Alphabetical,
		),
		lexer.Condition[variant](identifier, lexer.Any[variant](
			lexer.Alphanumeric,
			lexer.Equals[variant]('_'),
		)),
	)
	l.Load(string(src))
	return l, nil
}

type variant int
