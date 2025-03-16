package hdl

import (
	"fmt"
	"strconv"
	"strings"
)

type Statement interface {
	Literal() string
}

type Expression interface {
	Statement
}

type Parser struct {
	lexer lexer
}

func (p *Parser) Parse() (Chip, error) {
	if _, err := p.expect(chip); err != nil {
		return Chip{}, err
	}
	name, err := p.expect(identifier)
	if err != nil {
		return Chip{}, err
	}
	inputs, err := p.parseInputDefinition()
	if err != nil {
		return Chip{}, err
	}
	if _, err := p.expect(arrow); err != nil {
		return Chip{}, err
	}
	outputs, err := p.parseOutputDefinition()
	if err != nil {
		return Chip{}, err
	}
	body, err := p.parseStatementBlock()
	if err != nil {
		return Chip{}, err
	}
	return Chip{
		name:    name.literal,
		inputs:  inputs,
		outputs: outputs,
		body:    body,
	}, nil
}

func (p *Parser) parseInputDefinition() (map[string]int, error) {
	inputs := make(map[string]int)
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
		inputs[name.literal] = parsedSize
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inputs, nil
}

func (p *Parser) parseOutputDefinition() ([]int, error) {
	outputs := make([]int, 0)
	err := p.parseList(func() error {
		size, err := p.expect(integer)
		if err != nil {
			return err
		}
		parsedSize, err := strconv.Atoi(size.literal)
		if err != nil {
			return err
		}
		outputs = append(outputs, parsedSize)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outputs, nil
}

func (p *Parser) parseList(itemParser func() error) error {
	if _, err := p.expect(leftParenthesis); err != nil {
		return err
	}
	tok, err := p.lexer.peek()
	if err != nil {
		return err
	}
	for tok.variant != rightParenthesis {
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
	_, err = p.expect(rightParenthesis)
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
	case set:
		name, err := p.expect(identifier)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(equals); err != nil {
			return nil, err
		}
		expression, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return SetStatement{
			identifier: name.literal,
			expression: expression,
		}, nil
	case out:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return OutStatement{expression: expr}, nil
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
			literal: parsed,
		}, nil
	case identifier:
		next, err := p.lexer.peek()
		if err != nil {
			return nil, err
		}
		if next.variant != leftParenthesis {
			return IdentifierExpression{literal: tok.literal}, nil
		}
		args := make(map[string]Expression)
		err = p.parseList(func() error {
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
		})
		if err != nil {
			return nil, err
		}
		return CallExpression{
			name: tok.literal,
			args: args,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected token '%s'", tok.literal)
	}
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

type Chip struct {
	name    string
	inputs  map[string]int
	outputs []int
	body    []Statement
}

func (c Chip) Literal() string {
	inputs := make([]string, 0, len(c.inputs))
	for name, length := range c.inputs {
		inputs = append(inputs, fmt.Sprintf("%s: %d", name, length))
	}
	outputs := make([]string, 0, len(c.outputs))
	for _, output := range c.outputs {
		outputs = append(outputs, fmt.Sprintf("%d", output))
	}
	body := make([]string, 0, len(c.body))
	for _, stmt := range c.body {
		body = append(body, stmt.Literal())
	}
	return fmt.Sprintf(
		"chip %s (%s) -> (%s) { %s }",
		c.name,
		strings.Join(inputs, ", "),
		strings.Join(outputs, ", "),
		body,
	)
}

type SetStatement struct {
	identifier string
	expression Expression
}

func (s SetStatement) Literal() string {
	return fmt.Sprintf("set %s = %s", s.identifier, s.expression.Literal())
}

type OutStatement struct {
	expression Expression
}

func (o OutStatement) Literal() string {
	return fmt.Sprintf("out %s", o.expression.Literal())
}

type CallExpression struct {
	name string
	args map[string]Expression
}

func (c CallExpression) Literal() string {
	args := make([]string, 0, len(c.args))
	for name, expression := range c.args {
		args = append(args, fmt.Sprintf("%s: %s", name, expression.Literal()))
	}
	return fmt.Sprintf("%s(%s)", c.name, strings.Join(args, ","))
}

type IntegerExpression struct {
	literal int
}

func (i IntegerExpression) Literal() string {
	return fmt.Sprintf("%d", i.literal)
}

type IdentifierExpression struct {
	literal string
}

func (i IdentifierExpression) Literal() string {
	return i.literal
}
