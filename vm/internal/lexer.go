package internal

import (
	"github.com/crookdc/nand2tetris/lexer"
	"io"
)

var (
	NullToken Token = lexer.Token[Variant]{}
)

type Token = lexer.Token[Variant]

type Lexer = *lexer.Lexer[Variant]

const (
	Push Variant = iota
	Pop
	Add
	Sub
	Neg
	Eq
	Gt
	Lt
	And
	Or
	Not
	Constant
	Static
	Pointer
	Local
	Arg
	Tmp
	This
	That
	Integer
	Function
	Identifier
	Return
	Call
	Goto
	Label
	IfGoto
)

func NewLexer(r io.Reader) (*lexer.Lexer[Variant], error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	l := lexer.NewLexer[Variant](
		lexer.Params[Variant]{
			Symbols: make(map[uint8]Variant),
			Ignore: lexer.Any(
				lexer.Whitespace[Variant],
				lexer.LineComment[Variant]("//"),
			),
		},
		lexer.Integer[Variant](Integer),
		lexer.Keywords[Variant](
			map[string]Variant{
				"push":     Push,
				"pop":      Pop,
				"add":      Add,
				"sub":      Sub,
				"neg":      Neg,
				"and":      And,
				"or":       Or,
				"not":      Not,
				"eq":       Eq,
				"lt":       Lt,
				"gt":       Gt,
				"constant": Constant,
				"local":    Local,
				"static":   Static,
				"argument": Arg,
				"temp":     Tmp,
				"this":     This,
				"that":     That,
				"pointer":  Pointer,
				"function": Function,
				"return":   Return,
				"call":     Call,
				"goto":     Goto,
				"label":    Label,
				"if-goto":  IfGoto,
			},
			lexer.Any(lexer.Alphabetical, lexer.Equals[Variant]('-')),
		),
		lexer.Condition[Variant](Identifier, lexer.Any[Variant](
			lexer.Alphanumeric,
			lexer.Equals[Variant]('_'),
			lexer.Equals[Variant]('.'),
			lexer.Equals[Variant]('$'),
		)),
	)
	l.Load(string(src))
	return l, nil
}

type Variant int
