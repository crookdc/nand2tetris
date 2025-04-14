package asm

import (
	"errors"
	"fmt"
	"github.com/crookdc/nand2tetris/lexer"
	"io"
)

var (
	symbols = map[uint8]variant{
		'\n': linefeed,
		'@':  at,
		'-':  minus,
		'+':  plus,
		'&':  and,
		'|':  or,
		';':  semicolon,
		'(':  lparen,
		')':  rparen,
		'=':  equals,
		'!':  bang,
	}
	keywords = map[string]variant{
		"JGT": jgt,
		"JEQ": jeq,
		"JGE": jge,
		"JLT": jlt,
		"JNE": jne,
		"JLE": jle,
		"JMP": jmp,
	}
)

const (
	at variant = iota
	minus
	plus
	and
	or
	equals
	bang
	identifier
	integer
	lparen
	rparen
	semicolon
	jgt
	jeq
	jge
	jlt
	jne
	jle
	jmp
	linefeed
	comment
)

type variant int

func NewLexer() *lexer.Lexer[variant] {
	return lexer.NewLexer[variant](
		lexer.Params[variant]{
			Symbols: symbols,
			Ignore: lexer.Any(
				lexer.All(
					lexer.Whitespace,
					lexer.Not(lexer.Equals[variant]('\n')),
				),
				lexer.LineComment[variant]("//"),
			),
		},
		lexer.Integer[variant](integer),
		lexer.Keywords[variant](keywords, lexer.Alphabetical),
		lexer.Condition(identifier, lexer.Any(
			lexer.Alphanumeric,
			lexer.Equals[variant]('_'),
			lexer.Equals[variant]('.'),
			lexer.Equals[variant]('$'),
			lexer.Equals[variant](':'),
		)),
	)
}

func LoadedLexer(src string) *lexer.Lexer[variant] {
	lex := NewLexer()
	lex.Load(src)
	return lex
}

type instruction interface {
	Literal() string
}

type load struct {
	value lexer.Token[variant]
}

func (l load) Literal() string {
	return fmt.Sprintf("@%s", l.value.Literal)
}

type compute struct {
	dest *lexer.Token[variant]
	comp string
	jump *lexer.Token[variant]
}

func (c compute) Literal() string {
	var str string
	if c.dest != nil {
		str += fmt.Sprintf("%s=", c.dest.Literal)
	}
	str += c.comp
	if c.jump != nil {
		str += fmt.Sprintf(";%s", c.jump.Literal)
	}
	return str
}

type label struct {
	value lexer.Token[variant]
}

func (l label) Literal() string {
	return fmt.Sprintf("(%s)", l.value.Literal)
}

type parser struct {
	lexer *lexer.Lexer[variant]
}

func (p *parser) more() bool {
	return p.lexer.More()
}

func (p *parser) next() (instruction, error) {
	if err := p.seek(p.clear); err != nil {
		return nil, err
	}
	tok, err := p.lexer.Peek()
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	switch tok.Variant {
	case at:
		return p.a()
	case lparen:
		return p.label()
	default:
		return p.c()
	}
}

func (p *parser) a() (load, error) {
	if _, err := p.want(at); err != nil {
		return load{}, err
	}
	tok, err := p.lexer.Next()
	if err != nil {
		return load{}, err
	}
	if tok.Variant != integer && tok.Variant != identifier {
		return load{}, fmt.Errorf("unexpected token for A-instruction '%v'", tok)
	}
	if err := p.seek(p.clear); err != nil {
		return load{}, err
	}
	return load{value: tok}, nil
}

func (p *parser) label() (label, error) {
	if _, err := p.want(lparen); err != nil {
		return label{}, err
	}
	name, err := p.want(identifier)
	if err != nil {
		return label{}, err
	}
	if _, err := p.want(rparen); err != nil {
		return label{}, err
	}
	if err := p.seek(p.clear); err != nil {
		return label{}, err
	}
	return label{value: name}, nil
}

func (p *parser) c() (comp compute, err error) {
	tok, err := p.lexer.Next()
	if err != nil {
		return compute{}, err
	}
	next, err := p.lexer.Peek()
	if err != nil {
		return compute{}, err
	}
	if next.Variant == equals {
		_, _ = p.want(equals)
		comp.dest = &lexer.Token[variant]{
			Variant: tok.Variant,
			Literal: tok.Literal,
		}
		// Fetch the next token for parsing the compute field
		tok, err = p.lexer.Next()
		if err != nil {
			return compute{}, err
		}
	}
	for tok.Variant != semicolon && tok.Variant != linefeed {
		comp.comp += tok.Literal
		tok, err = p.lexer.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return compute{}, err
		}
	}
	if tok.Variant == semicolon {
		jmp, err := p.lexer.Next()
		if err != nil {
			return compute{}, err
		}
		comp.jump = &lexer.Token[variant]{
			Variant: jmp.Variant,
			Literal: jmp.Literal,
		}
		if err := p.seek(p.clear); err != nil {
			return compute{}, err
		}
	}
	return comp, nil
}

func (p *parser) seek(fn func(*lexer.Token[variant]) bool) error {
	tok, err := p.lexer.Peek()
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	for fn(&tok) {
		_, err = p.lexer.Next()
		if err != nil {
			return err
		}
		tok, err = p.lexer.Peek()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) clear(tok *lexer.Token[variant]) bool {
	return tok.Variant == linefeed || tok.Variant == comment
}

// want asserts that the next token supplied by the OldLexer is of a given variant. If the OldLexer returns a different
// variant than the one expected then an error is returned. If the expected token does appear then it is returned to the
// caller.
func (p *parser) want(v variant) (lexer.Token[variant], error) {
	tok, err := p.lexer.Next()
	if err != nil {
		return lexer.Token[variant]{}, err
	}
	if tok.Variant != v {
		return lexer.Token[variant]{}, fmt.Errorf("expected '%v' token but found '%v'", v, tok.Variant)
	}
	return tok, nil
}
