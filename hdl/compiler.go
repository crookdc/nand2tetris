package hdl

import (
	"errors"
	"fmt"
)

var (
	ErrChipNotFound              = errors.New("chip not found")
	ErrInvalidArgumentExpression = errors.New("invalid argument Expression")
	ErrInvalidArrayExpression    = errors.New("invalid array Expression")
)

func NAND(breadboard *Breadboard) (input ID, output ID) {
	output = breadboard.Allocate(1, nil)
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

func DFF(breadboard *Breadboard) (input ID, output ID) {
	output = breadboard.Allocate(1, nil)
	input = breadboard.Allocate(1, nil)
	clk := breadboard.Allocate(1, func(id ID, bytes []byte) {
		if bytes[0] == 1 {
			breadboard.Set(Pin{ID: output, Index: 0}, breadboard.Get(Pin{ID: input, Index: 0}))
		}
	})
	breadboard.Connect(Wire{
		Head: Pin{
			ID:    breadboard.CLK,
			Index: 0,
		},
		Tail: Pin{
			ID:    clk,
			Index: 0,
		},
	})
	return
}

type Chip struct {
	Environment map[string]ID
	Outputs     []ID
}

func Compile(breadboard *Breadboard, definition ChipStatement, support map[string]ChipStatement) (Chip, error) {
	inputs := make(map[string]ID)
	for name, size := range definition.Inputs {
		id := breadboard.Allocate(int(size), nil)
		inputs[name] = id
	}
	return compile(state{
		breadboard: breadboard,
		definition: definition,
		support:    support,
	}, inputs)
}

type state struct {
	breadboard *Breadboard
	definition ChipStatement
	support    map[string]ChipStatement
}

func compile(s state, inputs map[string]ID) (Chip, error) {
	ch := Chip{
		Environment: make(map[string]ID),
		Outputs:     make([]ID, len(s.definition.Outputs)),
	}
	for name, id := range inputs {
		ch.Environment[name] = id
	}
	for i, size := range s.definition.Outputs {
		id := s.breadboard.Allocate(int(size), nil)
		ch.Outputs[i] = id
	}
	var output int
	for _, statement := range s.definition.Body {
		switch stmt := statement.(type) {
		case OutStatement:
			ids, err := expression(s, &ch, stmt.Expression)
			if err != nil {
				return Chip{}, err
			}
			for _, id := range ids {
				if err := s.breadboard.ConnectGroup(id, ch.Outputs[output]); err != nil {
					return Chip{}, err
				}
				output++
			}
		case SetStatement:
			ids, err := expression(s, &ch, stmt.Expression)
			if err != nil {
				return Chip{}, err
			}
			for i, id := range ids {
				ident := stmt.Identifiers[i]
				if ident == "_" {
					continue
				}
				if _, ok := ch.Environment[ident]; ok {
					return Chip{}, fmt.Errorf("cannot redeclare identifier '%stmt'", ident)
				}
				ch.Environment[ident] = id
			}
		case UseStatement:
			continue
		default:
			return Chip{}, fmt.Errorf("unexpected statement '%stmt'", stmt.Literal())
		}
	}
	return ch, nil
}

func expression(s state, c *Chip, expr Expression) ([]ID, error) {
	switch e := expr.(type) {
	case CallExpression:
		return call(s, c, e)
	case IntegerExpression:
		if e.Integer == 0 {
			return []ID{s.breadboard.Zero}, nil
		}
		return []ID{s.breadboard.One}, nil
	case IndexedExpression:
		id, err := indexed(s, c, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case IdentifierExpression:
		id, err := identified(c, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case ArrayExpression:
		id, err := array(s, c, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	default:
		return nil, fmt.Errorf("invalid Expression '%s'", e.Literal())
	}
}

func call(s state, c *Chip, e CallExpression) ([]ID, error) {
	switch e.Name {
	case "feedback":
		return c.Outputs, nil
	case "nand":
		input, output := NAND(s.breadboard)
		in, err := expression(s, c, e.Args["in"])
		if err != nil {
			return nil, err
		}
		if len(in) != 1 {
			return nil, ErrInvalidArgumentExpression
		}
		if err := s.breadboard.ConnectGroup(in[0], input); err != nil {
			return nil, err
		}
		return []ID{output}, nil
	case "dff":
		input, output := DFF(s.breadboard)
		in, err := expression(s, c, e.Args["in"])
		if err != nil {
			return nil, err
		}
		if len(in) != 1 {
			return nil, ErrInvalidArgumentExpression
		}
		if err := s.breadboard.ConnectGroup(in[0], input); err != nil {
			return nil, err
		}
		return []ID{output}, nil
	default:
		definition, ok := s.support[e.Name]
		if !ok {
			return nil, ErrChipNotFound
		}
		params := make(map[string]ID)
		for arg, expr := range e.Args {
			val, err := expression(s, c, expr)
			if err != nil {
				return nil, err
			}
			if len(val) != 1 {
				return nil, ErrInvalidArgumentExpression
			}
			params[arg] = val[0]
		}
		s.definition = definition
		ch, err := compile(s, params)
		if err != nil {
			return nil, err
		}
		return ch.Outputs, nil
	}
}

func indexed(s state, c *Chip, e IndexedExpression) (ID, error) {
	head := c.Environment[e.Identifier]
	tail := s.breadboard.Allocate(1, nil)
	s.breadboard.Connect(Wire{
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

func identified(c *Chip, e IdentifierExpression) (ID, error) {
	head, ok := c.Environment[e.Identifier]
	if !ok {
		return 0, fmt.Errorf("invalid identifier '%s'", e.Identifier)
	}
	return head, nil
}

func array(s state, c *Chip, e ArrayExpression) (ID, error) {
	result := s.breadboard.Allocate(len(e.Values), nil)
	for i := range e.Values {
		head, err := expression(s, c, e.Values[i])
		if err != nil {
			return 0, err
		}
		if len(head) != 1 {
			return 0, ErrInvalidArrayExpression
		}
		s.breadboard.Connect(Wire{
			Head: Pin{
				ID:    head[0],
				Index: 0,
			},
			Tail: Pin{
				ID:    result,
				Index: i,
			},
		})
	}
	return result, nil
}
