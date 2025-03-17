package hdl

import (
	"errors"
	"fmt"
)

var (
	ErrChipDefinitionDoesNotExist = errors.New("chip definition does not exist")
	ErrUnsupportedExpression      = errors.New("unsupported expression")
)

type Builtin interface {
	Signal()
}

func NewPins(length byte) *Pins {
	return &Pins{
		pins: make([]*Pin, length),
	}
}

type Pins struct {
	pins []*Pin
}

func (p *Pins) Signal() {

}

type Constant struct {
	n byte
}

func (l *Constant) Signal() {}

type NAND struct {
	a      *Pin
	b      *Pin
	output *Pin
}

func (n *NAND) Signal() {
	if n.a.signal == 0 || n.b.signal == 0 {
		n.output.Set(1)
	} else {
		n.output.Set(0)
	}
}

type Pin struct {
	Binding Builtin
	signal  byte
}

func (p *Pin) Signal() {}

func (p *Pin) Set(signal byte) {
	if signal != 0 && signal != 1 {
		panic("tried to set invalid signal value")
	}
	if signal == p.signal {
		return
	}
	p.signal = signal
	if p.Binding != nil {
		p.Binding.Signal()
	}
}

type Chip struct {
	Name        string
	Inputs      map[string]*Pins
	Outputs     []*Pins
	invocations []Builtin
}

type Compiler struct {
	definitions map[string]ChipDefinition
}

func NewCompiler(definitions []ChipDefinition) *Compiler {
	c := &Compiler{
		definitions: make(map[string]ChipDefinition),
	}
	for _, definition := range definitions {
		c.definitions[definition.Name] = definition
	}
	return c
}

func (c *Compiler) CompileChip(name string) (Chip, error) {
	definition, ok := c.definitions[name]
	if !ok {
		return Chip{}, ErrChipDefinitionDoesNotExist
	}
	outputs := make([]*Pins, len(definition.Outputs))
	for i, output := range definition.Outputs {
		outputs[i] = NewPins(output)
	}
	ch := Chip{
		Name:    definition.Name,
		Inputs:  make(map[string]*Pins),
		Outputs: outputs,
	}
	for key, input := range definition.Inputs {
		ch.Inputs[key] = NewPins(input)
	}

	for _, _ = range definition.Body {
		/*switch v := stmt.(type) {
		case OutStatement:
			bn, err := c.compileExpression(&ch, v.expression)
			if err != nil {
				return Chip{}, err
			}
			ch.Outputs[idx]
		}
		*/
	}
	panic("not implemented")
}

func (c *Compiler) compileExpression(ch *Chip, exp Expression) ([]Builtin, error) {
	switch v := exp.(type) {
	case IntegerExpression:
		return []Builtin{&Constant{n: byte(v.Integer)}}, nil
	case IndexedExpression:
		r, ok := ch.Inputs[v.Identifier]
		if !ok {
			return nil, fmt.Errorf("unknown identifier %s.%d", v.Identifier, v.Index)
		}
		return []Builtin{r.pins[v.Index]}, nil
	case CallExpression:
		panic("not implemented")
	default:
		return nil, ErrUnsupportedExpression
	}
}
