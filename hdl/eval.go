package hdl

import (
	"errors"
	"fmt"
)

var (
	ErrChipNotFound              = errors.New("chip not found")
	ErrInvalidArgumentExpression = errors.New("invalid argument expression")
	ErrInvalidArrayExpression    = errors.New("invalid array expression")
)

func NAND(breadboard *Breadboard) (input ID, output ID) {
	output = breadboard.Allocate(1, Noop)
	input = breadboard.Allocate(2, func(id ID, bytes []byte) {
		if bytes[0] == 0 || bytes[1] == 0 {
			breadboard.Set(Pin{
				ID:    output,
				Index: 0,
			}, 1)
		} else {
			breadboard.Set(Pin{
				ID:    output,
				Index: 0,
			}, 0)
		}
	})
	return
}

type Chip struct {
	Inputs  map[string]ID
	Outputs []ID
}

func NewCompiler(support map[string]ChipDefinition) *Compiler {
	return &Compiler{
		breadboard: NewBreadboard(),
		support:    support,
	}
}

type Compiler struct {
	breadboard *Breadboard
	support    map[string]ChipDefinition
}

func (c *Compiler) Compile(main ChipDefinition) (Chip, error) {
	compiled := Chip{
		Inputs:  make(map[string]ID),
		Outputs: make([]ID, len(main.Outputs)),
	}
	for param, size := range main.Inputs {
		id := c.breadboard.Allocate(int(size), Noop)
		compiled.Inputs[param] = id
	}
	for i, size := range main.Outputs {
		id := c.breadboard.Allocate(int(size), Noop)
		compiled.Outputs[i] = id
	}
	var counter int
	for _, statement := range main.Body {
		switch s := statement.(type) {
		case OutStatement:
			ids, err := c.expression(&compiled, s.expression)
			if err != nil {
				return Chip{}, err
			}
			for _, id := range ids {
				compiled.Outputs[counter] = id
				counter++
			}
		default:
			panic("not implemented")
		}
	}
	return compiled, nil
}

func (c *Compiler) expression(chip *Chip, exp Expression) ([]ID, error) {
	switch e := exp.(type) {
	case CallExpression:
		return c.evaluateCallExpression(chip, e)
	case IntegerExpression:
		id := c.breadboard.Allocate(1, Noop)
		c.breadboard.Set(Pin{ID: id, Index: 0}, byte(e.Integer))
		return []ID{id}, nil
	case IndexedExpression:
		id, err := c.evaluateIndexedExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case IdentifierExpression:
		id, err := c.evaluateIdentifierExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case ArrayExpression:
		id, err := c.evaluateArrayExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	default:
		return nil, fmt.Errorf("invalid expression '%s'", e.Literal())
	}
}

func (c *Compiler) evaluateCallExpression(chip *Chip, e CallExpression) ([]ID, error) {
	if e.Name == "nand" {
		return c.evaluateNandChipInvocation(chip, e)
	}
	return c.evaluateSupportChipInvocation(chip, e)
}

func (c *Compiler) evaluateNandChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	input, output := NAND(c.breadboard)
	in, err := c.expression(chip, e.Args["in"])
	if err != nil {
		return nil, err
	}
	if len(in) != 1 {
		return nil, ErrInvalidArgumentExpression
	}
	if err := c.breadboard.ConnectGroup(in[0], input); err != nil {
		return nil, err
	}
	return []ID{output}, nil
}

func (c *Compiler) evaluateSupportChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	definition, ok := c.support[e.Name]
	if !ok {
		return nil, ErrChipNotFound
	}
	ch, err := c.Compile(definition)
	if err != nil {
		return nil, err
	}
	for arg, valExpr := range e.Args {
		val, err := c.expression(chip, valExpr)
		if err != nil {
			return nil, err
		}
		if len(val) != 1 {
			return nil, ErrInvalidArgumentExpression
		}
		if err := c.breadboard.ConnectGroup(val[0], ch.Inputs[arg]); err != nil {
			return nil, err
		}
	}
	return ch.Outputs, nil
}

func (c *Compiler) evaluateIndexedExpression(chip *Chip, e IndexedExpression) (ID, error) {
	head := chip.Inputs[e.Identifier]
	tail := c.breadboard.Allocate(1, Noop)
	c.breadboard.Connect(Wire{
		Head: Pin{
			ID:    head,
			Index: e.Index,
		},
		Tail: Pin{
			ID:    tail,
			Index: 0,
		},
	})
	return tail, nil
}

func (c *Compiler) evaluateIdentifierExpression(chip *Chip, e IdentifierExpression) (ID, error) {
	head := chip.Inputs[e.Identifier]
	return head, nil
}

func (c *Compiler) evaluateArrayExpression(chip *Chip, e ArrayExpression) (ID, error) {
	array := c.breadboard.Allocate(len(e.Values), Noop)
	for i := range e.Values {
		head, err := c.expression(chip, e.Values[i])
		if err != nil {
			return 0, err
		}
		if len(head) != 1 {
			return 0, ErrInvalidArrayExpression
		}
		c.breadboard.Connect(Wire{
			Head: Pin{
				ID:    head[0],
				Index: 0,
			},
			Tail: Pin{
				ID:    array,
				Index: i,
			},
		})
	}
	return array, nil
}
