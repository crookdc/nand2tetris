package vm

import (
	"fmt"
	"github.com/crookdc/nand2tetris/lexer"
	"io"
	"strconv"
)

const (
	push variant = iota
	constant
	local
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
			},
			lexer.Alphabetical,
		),
	)
	l.Load(string(src))
	return l, nil
}

type variant int

func Evaluate(r io.Reader) ([]string, error) {
	l, err := newLexer(r)
	if err != nil {
		return nil, err
	}
	assembly := make([]string, 0)
	for l.More() {
		next, err := l.Peek()
		if err != nil {
			return nil, err
		}
		switch next.Variant {
		case push:
			asm, err := parsePushCommand(l)
			if err != nil {
				return nil, err
			}
			assembly = append(assembly, asm...)
		default:
			panic(fmt.Errorf("variant %v is not yet supported", next.Variant))
		}
	}
	return assembly, nil
}

func expect(l *lexer.Lexer[variant], expected variant) (lexer.Token[variant], error) {
	token, err := l.Next()
	if err != nil {
		return lexer.Token[variant]{}, err
	}
	if token.Variant != expected {
		return lexer.Token[variant]{}, fmt.Errorf("expected token variant %v but got %v", expected, token.Variant)
	}
	return token, nil
}

func parsePushCommand(l *lexer.Lexer[variant]) ([]string, error) {
	if _, err := expect(l, push); err != nil {
		return nil, err
	}
	read, err := parseSegmentRead(l)
	if err != nil {
		return nil, err
	}
	return append(
		read,
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",
	), nil
}

func parseSegmentRead(l *lexer.Lexer[variant]) ([]string, error) {
	token, err := l.Next()
	if err != nil {
		return nil, err
	}
	switch token.Variant {
	case constant:
		value, err := l.Next()
		if err != nil {
			return nil, err
		}
		if value.Variant != integer {
			return nil, fmt.Errorf("invalid integer token %v", value.Literal)
		}
		parsed, err := strconv.Atoi(value.Literal)
		if err != nil || parsed < 0 || parsed > 32768 {
			return nil, fmt.Errorf("invalid constant value %v", value.Literal)
		}
		return []string{
			fmt.Sprintf("@%d", parsed),
			"D=A",
		}, nil
	case local:
		value, err := l.Next()
		if err != nil {
			return nil, err
		}
		if value.Variant != integer {
			return nil, fmt.Errorf("invalid integer token %v", value.Literal)
		}
		parsed, err := strconv.Atoi(value.Literal)
		if err != nil || parsed < 0 {
			return nil, fmt.Errorf("invalid static index %v", value.Literal)
		}
		return []string{
			"@LCL",
			"D=M",
			fmt.Sprintf("@%d", parsed),
			"A=D+A",
			"D=M",
		}, nil
	default:
		panic(fmt.Errorf("push type variant %v is not yet supported", token.Variant))
	}
}
