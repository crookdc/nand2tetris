package vm

import (
	"github.com/crookdc/nand2tetris/lexer"
	"io"
)

const (
	push variant = iota
	constant
	local
	arg
	this
	that
	integer
)

func newLexer(r io.Reader) (*lexer.Lexer[variant], error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	l := lexer.NewLexer[variant](
		lexer.Params[variant]{
			Symbols: make(map[uint8]variant),
			Ignore:  lexer.Whitespace,
		},
		lexer.Integer[variant](integer),
		lexer.Keywords[variant](
			map[string]variant{
				"push":     push,
				"constant": constant,
				"local":    local,
				"arg":      arg,
				"this":     this,
				"that":     that,
			},
			lexer.Alphabetical,
		),
	)
	l.Load(string(src))
	return l, nil
}

type variant int
