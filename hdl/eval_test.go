package hdl

import (
	"testing"
)

func TestNAND(t *testing.T) {
	breadboard := NewBreadboard()
	input, output := NAND(breadboard)
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 1,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 0 {
		t.Errorf("expected output to be 0 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 0)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
}

func TestDFF(t *testing.T) {
	breadboard := NewBreadboard()
	load, input, output := DFF(breadboard)
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 0 {
		t.Errorf("expected output to be 0 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    load,
		Index: 0,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 0)
	breadboard.Set(Pin{
		ID:    load,
		Index: 0,
	}, 0)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    load,
		Index: 0,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 0 {
		t.Errorf("expected output to be 0 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
}

func TestEvaluator_Evaluate(t *testing.T) {
	compiler := NewEvaluator(map[string]ChipDefinition{
		"NOT": {
			Name: "NOT",
			Inputs: map[string]byte{
				"a": 1,
			},
			Outputs: []byte{1},
			Body: []Statement{
				OutStatement{
					Expression: CallExpression{
						Name: "nand",
						Args: map[string]Expression{
							"in": ArrayExpression{
								Values: []Expression{
									IndexedExpression{
										Identifier: "a",
										Index:      0,
									},
									IntegerExpression{Integer: 1},
								},
							},
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
					Expression: CallExpression{
						Name: "NOT",
						Args: map[string]Expression{
							"a": CallExpression{
								Name: "nand",
								Args: map[string]Expression{
									"in": ArrayExpression{
										Values: []Expression{
											IndexedExpression{
												Identifier: "input",
												Index:      0,
											},
											IndexedExpression{
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
			},
		},
	})
	compiled, err := compiler.Evaluate(ChipDefinition{
		Name: "TST",
		Inputs: map[string]byte{
			"in": 1,
		},
		Outputs: []byte{1},
		Body: []Statement{
			OutStatement{
				Expression: CallExpression{
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

	compiler.Breadboard.Set(Pin{
		ID:    compiled.Environment["in"],
		Index: 0,
	}, 1)
	Tick(compiler.Breadboard)
	if compiler.Breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 1 {
		t.Errorf("expected TST 1 to equal 1")
	}

	compiler.Breadboard.Set(Pin{
		ID:    compiled.Environment["in"],
		Index: 0,
	}, 0)
	Tick(compiler.Breadboard)
	if compiler.Breadboard.Get(Pin{ID: compiled.Outputs[0], Index: 0}) != 0 {
		t.Errorf("expected TST 0 to equal 0")
	}
}
