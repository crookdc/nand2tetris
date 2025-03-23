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
	compiled := Chip{
		Inputs:  make(map[string]ID),
		Outputs: make([]ID, len(main.Outputs)),
	}
	for param, size := range main.Inputs {
		id := ev.Breadboard.Allocate(int(size), Noop)
		compiled.Inputs[param] = id
	}
	for i, size := range main.Outputs {
		id := ev.Breadboard.Allocate(int(size), Noop)
		compiled.Outputs[i] = id
	}
	var counter int
	for _, statement := range main.Body {
		switch s := statement.(type) {
		case OutStatement:
			ids, err := ev.expression(&compiled, s.Expression)
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

func (ev *Evaluator) expression(chip *Chip, exp Expression) ([]ID, error) {
	switch e := exp.(type) {
	case CallExpression:
		return ev.evaluateCallExpression(chip, e)
	case IntegerExpression:
		id := ev.Breadboard.Allocate(1, Noop)
		ev.Breadboard.Set(Pin{ID: id, Index: 0}, byte(e.Integer))
		return []ID{id}, nil
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
	if e.Name == "nand" {
		return ev.evaluateNandChipInvocation(chip, e)
	}
	return ev.evaluateSupportChipInvocation(chip, e)
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

func (ev *Evaluator) evaluateSupportChipInvocation(chip *Chip, e CallExpression) ([]ID, error) {
	definition, ok := ev.support[e.Name]
	if !ok {
		return nil, ErrChipNotFound
	}
	ch, err := ev.Evaluate(definition)
	if err != nil {
		return nil, err
	}
	for arg, valExpr := range e.Args {
		val, err := ev.expression(chip, valExpr)
		if err != nil {
			return nil, err
		}
		if len(val) != 1 {
			return nil, ErrInvalidArgumentExpression
		}
		if err := ev.Breadboard.ConnectGroup(val[0], ch.Inputs[arg]); err != nil {
			return nil, err
		}
	}
	return ch.Outputs, nil
}

func (ev *Evaluator) evaluateIndexedExpression(chip *Chip, e IndexedExpression) (ID, error) {
	head := chip.Inputs[e.Identifier]
	tail := ev.Breadboard.Allocate(1, Noop)
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
	head := chip.Inputs[e.Identifier]
	return head, nil
}

func (ev *Evaluator) evaluateArrayExpression(chip *Chip, e ArrayExpression) (ID, error) {
	array := ev.Breadboard.Allocate(len(e.Values), Noop)
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
