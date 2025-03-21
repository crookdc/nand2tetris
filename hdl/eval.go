package hdl

import "fmt"

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
			id, err := c.expression(&compiled, s.expression)
			if err != nil {
				return Chip{}, err
			}
			compiled.Outputs[counter] = id
			counter++
		default:
			panic("not implemented")
		}
	}
	return compiled, nil
}

func (c *Compiler) expression(chip *Chip, exp Expression) (ID, error) {
	switch e := exp.(type) {
	case CallExpression:
		if e.Name == "NAND" {
			input, output := NAND(c.breadboard)
			a, err := c.expression(chip, e.Args["a"])
			if err != nil {
				return -1, err
			}
			b, err := c.expression(chip, e.Args["b"])
			if err != nil {
				return -1, err
			}
			c.breadboard.Connect(Wire{
				Head: Pin{
					ID:    a,
					Index: 0,
				},
				Tail: Pin{
					ID:    input,
					Index: 0,
				},
			})
			c.breadboard.Connect(Wire{
				Head: Pin{
					ID:    b,
					Index: 0,
				},
				Tail: Pin{
					ID:    input,
					Index: 1,
				},
			})
			return output, nil
		}
		panic("not implemented")
	case IntegerExpression:
		id := c.breadboard.Allocate(1, Noop)
		c.breadboard.Set(Pin{ID: id, Index: 0}, byte(e.Integer))
		return id, nil
	case IndexedExpression:
		// Initially an identifier must be an input parameter, it cannot be a parameter which has been defined using the
		// `set` keyword.
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
	default:
		return -1, fmt.Errorf("invalid expression '%s'", e.Literal())
	}
}
