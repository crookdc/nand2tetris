package hdl

import "testing"

func TestNAND(t *testing.T) {
	breadboard := NewBreadboard()
	input, output := NAND(breadboard)
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 1)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 1,
	}, 1)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 0 {
		t.Errorf("expected output to be 0 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 0)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
}

func TestCompile(t *testing.T) {
	compiler := NewCompiler(nil)
	compiled, err := compiler.Compile(ChipDefinition{
		Name: "NOT",
		Inputs: map[string]byte{
			"a": 1,
		},
		Outputs: []byte{1},
		Body: []Statement{
			OutStatement{
				expression: CallExpression{
					Name: "NAND",
					Args: map[string]Expression{
						"a": IndexedExpression{
							Identifier: "a",
							Index:      0,
						},
						"b": IntegerExpression{Integer: 1},
					},
				},
			},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	compiler.breadboard.Set(Pin{
		ID:    compiled.Inputs["a"],
		Index: 0,
	}, 1)
	if compiler.breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 0 {
		t.Errorf("expected NOT(a: 1) to equal 0")
	}

	compiler.breadboard.Set(Pin{
		ID:    compiled.Inputs["a"],
		Index: 0,
	}, 0)
	if compiler.breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 1 {
		t.Errorf("expected NOT(a: 0) to equal 1")
	}
}
