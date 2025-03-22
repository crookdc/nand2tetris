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
	compiler := NewCompiler(map[string]ChipDefinition{
		"NOT": {
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
		},
		"AND": {
			Name: "AND",
			Inputs: map[string]byte{
				"input": 2,
			},
			Outputs: []byte{1},
			Body: []Statement{
				OutStatement{
					expression: CallExpression{
						Name: "NOT",
						Args: map[string]Expression{
							"a": CallExpression{
								Name: "NAND",
								Args: map[string]Expression{
									"a": IndexedExpression{
										Identifier: "input",
										Index:      0,
									},
									"b": IndexedExpression{
										Identifier: "input",
										Index:      1,
									},
								},
							},
						},
					},
				},
			},
		},
	})
	compiled, err := compiler.Compile(ChipDefinition{
		Name: "TST",
		Inputs: map[string]byte{
			"in": 1,
		},
		Outputs: []byte{1},
		Body: []Statement{
			OutStatement{
				expression: CallExpression{
					Name: "AND",
					Args: map[string]Expression{
						"input": ArrayExpression{
							Values: []Expression{
								IndexedExpression{
									Identifier: "input",
									Index:      0,
								},
								IntegerExpression{Integer: 1},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	compiler.breadboard.Set(Pin{
		ID:    compiled.Inputs["in"],
		Index: 0,
	}, 1)
	if compiler.breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 1 {
		t.Errorf("expected TST 1 to equal 1")
	}

	compiler.breadboard.Set(Pin{
		ID:    compiled.Inputs["in"],
		Index: 0,
	}, 0)
	if compiler.breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 0 {
		t.Errorf("expected TST 0 to equal 0")
	}
}
