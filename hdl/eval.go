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

func NewEvaluator(support map[string]ChipDefinition) *Evaluator {
	return &Evaluator{
		Breadboard: NewBreadboard(),
		support:    support,
	}
}

type Evaluator struct {
	Breadboard *Breadboard
	support    map[string]ChipDefinition
}

func (ev *Evaluator) Evaluate(main ChipDefinition) (Chip, error) {
	params := make(map[string]ID)
	for param, size := range main.Inputs {
		id := ev.Breadboard.Allocate(int(size), nil)
		params[param] = id
	}
	return ev.evaluateChip(main, params)
}

func (ev *Evaluator) evaluateChip(definition ChipDefinition, params map[string]ID) (Chip, error) {
	ch := Chip{
		Environment: make(map[string]ID),
		Outputs:     make([]ID, len(definition.Outputs)),
	}
	for param, id := range params {
		ch.Environment[param] = id
	}
	for i, size := range definition.Outputs {
		id := ev.Breadboard.Allocate(int(size), nil)
		ch.Outputs[i] = id
	}
	var output int
	for _, statement := range definition.Body {
		switch s := statement.(type) {
		case OutStatement:
			ids, err := ev.expression(&ch, s.Expression)
			if err != nil {
				return Chip{}, err
			}
			for _, id := range ids {
				if err := ev.Breadboard.ConnectGroup(id, ch.Outputs[output]); err != nil {
					return Chip{}, err
				}
				output++
			}
		case SetStatement:
			ids, err := ev.expression(&ch, s.Expression)
			if err != nil {
				return Chip{}, err
			}
			for i, id := range ids {
				ident := s.Identifiers[i]
				if ident == "_" {
					continue
				}
				if _, ok := ch.Environment[ident]; ok {
					return Chip{}, fmt.Errorf("cannot redeclare identifier '%s'", ident)
				}
				ch.Environment[ident] = id
			}
		default:
			return Chip{}, fmt.Errorf("unexpected statement '%s'", s.Literal())
		}
	}
	return ch, nil
}

func (ev *Evaluator) expression(chip *Chip, exp Expression) ([]ID, error) {
	switch e := exp.(type) {
	case CallExpression:
		return ev.evaluateCallExpression(chip, e)
	case IntegerExpression:
		if e.Integer == 0 {
			return []ID{ev.Breadboard.Zero}, nil
		}
		return []ID{ev.Breadboard.One}, nil
	case IndexedExpression:
		id, err := ev.evaluateIndexedExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case IdentifierExpression:
		id, err := ev.evaluateIdentifierExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	case ArrayExpression:
		id, err := ev.evaluateArrayExpression(chip, e)
		if err != nil {
			return nil, err
		}
		return []ID{id}, nil
	default:
		return nil, fmt.Errorf("invalid Expression '%s'", e.Literal())
	}
}

func (ev *Evaluator) evaluateCallExpression(chip *Chip, e CallExpression) ([]ID, error) {
	switch e.Name {
	case "feedback":
		return chip.Outputs, nil
	case "nand":
		return ev.evaluateNandChipInvocation(chip, e)
	case "dff":
		return ev.evaluateDffChipInvocation(chip, e)
	default:
		return ev.evaluateSupportChipInvocation(chip, e)
	}
}

func (ev *Evaluator) evaluateNandChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	input, output := NAND(ev.Breadboard)
	in, err := ev.expression(chip, e.Args["in"])
	if err != nil {
		return nil, err
	}
	if len(in) != 1 {
		return nil, ErrInvalidArgumentExpression
	}
	if err := ev.Breadboard.ConnectGroup(in[0], input); err != nil {
		return nil, err
	}
	return []ID{output}, nil
}

func (ev *Evaluator) evaluateDffChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	input, output := DFF(ev.Breadboard)
	in, err := ev.expression(chip, e.Args["in"])
	if err != nil {
		return nil, err
	}
	if len(in) != 1 {
		return nil, ErrInvalidArgumentExpression
	}
	if err := ev.Breadboard.ConnectGroup(in[0], input); err != nil {
		return nil, err
	}
	return []ID{output}, nil
}

func (ev *Evaluator) evaluateSupportChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	definition, ok := ev.support[e.Name]
	if !ok {
		return nil, ErrChipNotFound
	}
	params := make(map[string]ID)
	for arg, expr := range e.Args {
		val, err := ev.expression(chip, expr)
		if err != nil {
			return nil, err
		}
		if len(val) != 1 {
			return nil, ErrInvalidArgumentExpression
		}
		params[arg] = val[0]
	}
	ch, err := ev.evaluateChip(definition, params)
	if err != nil {
		return nil, err
	}
	return ch.Outputs, nil
}

func (ev *Evaluator) evaluateIndexedExpression(chip *Chip, e IndexedExpression) (ID, error) {
	head := chip.Environment[e.Identifier]
	tail := ev.Breadboard.Allocate(1, nil)
	ev.Breadboard.Connect(Wire{
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

func (ev *Evaluator) evaluateIdentifierExpression(chip *Chip, e IdentifierExpression) (ID, error) {
	head, ok := chip.Environment[e.Identifier]
	if !ok {
		return 0, fmt.Errorf("invalid identifier '%s'", e.Identifier)
	}
	return head, nil
}

func (ev *Evaluator) evaluateArrayExpression(chip *Chip, e ArrayExpression) (ID, error) {
	array := ev.Breadboard.Allocate(len(e.Values), nil)
	for i := range e.Values {
		head, err := ev.expression(chip, e.Values[i])
		if err != nil {
			return 0, err
		}
		if len(head) != 1 {
			return 0, ErrInvalidArrayExpression
		}
		ev.Breadboard.Connect(Wire{
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
