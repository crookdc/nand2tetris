package hdl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseFile(filename string) (map[string]ChipStatement, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	p := NewParser(Lexer{Source: string(src)})
	c := make(map[string]ChipStatement)
	stmts, err := p.Parse()
	if err != nil {
		return nil, err
	}
	for _, s := range stmts {
		switch t := s.(type) {
		case UseStatement:
			imported, err := ParseFile(filepath.Join(filepath.Dir(filename), t.FileName))
			if err != nil {
				return nil, err
			}
			for name, definition := range imported {
				c[name] = definition
			}
		case ChipStatement:
			c[t.Name] = t
		}
	}
	return c, nil
}

type Statement interface {
	Literal() string
}

type Expression interface {
	Statement
}

func NewParser(lexer Lexer) Parser {
	return Parser{
		lexer: lexer,
	}
}

type Parser struct {
	lexer Lexer
}

func (p *Parser) Parse() ([]Statement, error) {
	tok, err := p.lexer.peek()
	if err != nil {
		return nil, err
	}
	stmts := make([]Statement, 0)
	for tok.variant != eof {
		ch, err := p.parse()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, ch)
		tok, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}
	}
	return stmts, nil
}

func (p *Parser) parse() (Statement, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return nil, err
	}
	switch tok.variant {
	case chip:
		return p.parseChipStatement()
	case use:
		filename, err := p.expect(stringLiteral)
		if err != nil {
			return nil, err
		}
		return UseStatement{FileName: filename.literal}, nil
	default:
		return nil, fmt.Errorf("unexpected token '%s'", tok.literal)
	}
}

func (p *Parser) parseChipStatement() (ChipStatement, error) {
	name, err := p.expect(identifier)
	if err != nil {
		return ChipStatement{}, err
	}
	inputs, err := p.parseInputDefinition()
	if err != nil {
		return ChipStatement{}, err
	}
	if _, err := p.expect(arrow); err != nil {
		return ChipStatement{}, err
	}
	outputs, err := p.parseOutputDefinition()
	if err != nil {
		return ChipStatement{}, err
	}
	body, err := p.parseStatementBlock()
	if err != nil {
		return ChipStatement{}, err
	}
	return ChipStatement{
		Name:    name.literal,
		Inputs:  inputs,
		Outputs: outputs,
		Body:    body,
	}, nil
}

func (p *Parser) parseInputDefinition() (map[string]byte, error) {
	if _, err := p.expect(leftParenthesis); err != nil {
		return nil, err
	}
	inputs := make(map[string]byte)
	err := p.parseList(func() error {
		name, err := p.expect(identifier)
		if err != nil {
			return err
		}
		if _, err := p.expect(colon); err != nil {
			return err
		}
		size, err := p.expect(integer)
		if err != nil {
			return err
		}
		parsedSize, err := strconv.Atoi(size.literal)
		if err != nil {
			return err
		}
		inputs[name.literal] = byte(parsedSize)
		return nil
	}, rightParenthesis)
	if err != nil {
		return nil, err
	}
	return inputs, nil
}

func (p *Parser) parseOutputDefinition() ([]byte, error) {
	outputs := make([]byte, 0)
	if _, err := p.expect(leftParenthesis); err != nil {
		return nil, err
	}
	err := p.parseList(func() error {
		size, err := p.expect(integer)
		if err != nil {
			return err
		}
		parsedSize, err := strconv.Atoi(size.literal)
		if err != nil {
			return err
		}
		outputs = append(outputs, byte(parsedSize))
		return nil
	}, rightParenthesis)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}

func (p *Parser) parseList(itemParser func() error, terminator variant) error {
	tok, err := p.lexer.peek()
	if err != nil {
		return err
	}
	for tok.variant != terminator {
		if err := itemParser(); err != nil {
			return err
		}
		tok, err = p.lexer.peek()
		if err != nil {
			return err
		}
		if tok.variant == comma {
			tok, err = p.lexer.next()
			if err != nil {
				return err
			}
		}
	}
	_, err = p.expect(terminator)
	return err
}

func (p *Parser) parseStatementBlock() ([]Statement, error) {
	if _, err := p.expect(leftCurlyBrace); err != nil {
		return nil, err
	}
	tok, err := p.lexer.peek()
	if err != nil {
		return nil, err
	}
	statements := make([]Statement, 0)
	for tok.variant != rightCurlyBrace {
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
		tok, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}
	}
	if _, err := p.expect(rightCurlyBrace); err != nil {
		return nil, err
	}
	return statements, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return nil, err
	}
	switch tok.variant {
	case out:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return OutStatement{Expression: expr}, nil
	case set:
		identifiers := make([]string, 0)
		err := p.parseList(func() error {
			i, err := p.expect(identifier)
			if err != nil {
				return err
			}
			identifiers = append(identifiers, i.literal)
			return nil
		}, equals)
		if err != nil {
			return nil, err
		}
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return SetStatement{
			Identifiers: identifiers,
			Expression:  expr,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected token '%s'", tok.literal)
	}
}

func (p *Parser) parseExpression() (Expression, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return nil, err
	}
	switch tok.variant {
	case integer:
		parsed, err := strconv.Atoi(tok.literal)
		if err != nil {
			return nil, err
		}
		return IntegerExpression{
			Integer: parsed,
		}, nil
	case identifier:
		next, err := p.lexer.peek()
		if err != nil {
			return nil, err
		}
		if next.variant == dot {
			return p.parseIndexedExpression(tok)
		}
		if next.variant == leftParenthesis {
			return p.parseCallExpression(tok)
		}
		return IdentifierExpression{Identifier: tok.literal}, nil
	case leftBracket:
		values := make([]Expression, 0)
		err := p.parseList(func() error {
			value, err := p.parseExpression()
			if err != nil {
				return err
			}
			values = append(values, value)
			return nil
		}, rightBracket)
		if err != nil {
			return nil, err
		}
		return ArrayExpression{Values: values}, nil
	default:
		return nil, fmt.Errorf("unexpected token '%s'", tok.literal)
	}
}

func (p *Parser) parseIndexedExpression(ident token) (IndexedExpression, error) {
	_, _ = p.expect(dot)
	idx, err := p.expect(integer)
	if err != nil {
		return IndexedExpression{}, err
	}
	parsed, err := strconv.Atoi(idx.literal)
	if err != nil {
		return IndexedExpression{}, err
	}
	return IndexedExpression{Index: parsed, Identifier: ident.literal}, nil
}

func (p *Parser) parseCallExpression(ident token) (CallExpression, error) {
	args := make(map[string]Expression)
	if _, err := p.expect(leftParenthesis); err != nil {
		return CallExpression{}, err
	}
	err := p.parseList(func() error {
		name, err := p.expect(identifier)
		if err != nil {
			return err
		}
		if _, err := p.expect(colon); err != nil {
			return err
		}
		expr, err := p.parseExpression()
		if err != nil {
			return err
		}
		args[name.literal] = expr
		return nil
	}, rightParenthesis)
	if err != nil {
		return CallExpression{}, err
	}
	return CallExpression{
		Name: ident.literal,
		Args: args,
	}, nil
}

func (p *Parser) expect(v variant) (token, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return token{}, err
	}
	if tok.variant != v {
		return token{}, fmt.Errorf("unexpected token '%s'", tok.literal)
	}
	return tok, nil
}

type ChipStatement struct {
	Name    string
	Inputs  map[string]byte
	Outputs []byte
	Body    []Statement
}

func (c ChipStatement) Literal() string {
	inputs := make([]string, 0, len(c.Inputs))
	for name, length := range c.Inputs {
		inputs = append(inputs, fmt.Sprintf("%s: %d", name, length))
	}
	outputs := make([]string, 0, len(c.Outputs))
	for _, output := range c.Outputs {
		outputs = append(outputs, fmt.Sprintf("%d", output))
	}
	body := make([]string, 0, len(c.Body))
	for _, stmt := range c.Body {
		body = append(body, stmt.Literal())
	}
	return fmt.Sprintf(
		"chip %s (%s) -> (%s) { %s }",
		c.Name,
		strings.Join(inputs, ", "),
		strings.Join(outputs, ", "),
		body,
	)
}

type OutStatement struct {
	Expression Expression
}

func (o OutStatement) Literal() string {
	return fmt.Sprintf("out %s", o.Expression.Literal())
}

type SetStatement struct {
	Identifiers []string
	Expression  Expression
}

func (s SetStatement) Literal() string {
	return fmt.Sprintf("set %s = %s", strings.Join(s.Identifiers, ", "), s.Expression.Literal())
}

type UseStatement struct {
	FileName string
}

func (u UseStatement) Literal() string {
	return fmt.Sprintf("use \"%s\"", u.FileName)
}

type CallExpression struct {
	Name string
	Args map[string]Expression
}

func (c CallExpression) Literal() string {
	args := make([]string, 0, len(c.Args))
	for name, expression := range c.Args {
		args = append(args, fmt.Sprintf("%s: %s", name, expression.Literal()))
	}
	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(args, ","))
}

type IntegerExpression struct {
	Integer int
}

func (i IntegerExpression) Literal() string {
	return fmt.Sprintf("%d", i.Integer)
}

type IndexedExpression struct {
	Identifier string
	Index      int
}

func (i IndexedExpression) Literal() string {
	return fmt.Sprintf("%s.%d", i.Identifier, i.Index)
}

type ArrayExpression struct {
	Values []Expression
}

func (a ArrayExpression) Literal() string {
	values := make([]string, len(a.Values))
	for i, value := range a.Values {
		values[i] = value.Literal()
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ","))
}

type IdentifierExpression struct {
	Identifier string
}

func (i IdentifierExpression) Literal() string {
	return i.Identifier
}
