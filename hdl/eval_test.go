package hdl

import (
	"fmt"
	"reflect"
	"testing"
)

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
					expression: CallExpression{
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

func TestAnd16(t *testing.T) {
	source := `
	chip not (in: 1) -> (1) {
		out nand(in: [in.0, 1])
	}
	
	chip and (in: 2) -> (1) {
		out not(in: nand(in: [in.0, in.1]))
	}
	
	chip and_16 (a: 16, b: 16) -> (16) {
		out [
			and(in: [a.0, b.0]),
			and(in: [a.1, b.1]),
			and(in: [a.2, b.2]),
			and(in: [a.3, b.3]),
			and(in: [a.4, b.4]),
			and(in: [a.5, b.5]),
			and(in: [a.6, b.6]),
			and(in: [a.7, b.7]),
			and(in: [a.8, b.8]),
			and(in: [a.9, b.9]),
			and(in: [a.10, b.10]),
			and(in: [a.11, b.11]),
			and(in: [a.12, b.12]),
			and(in: [a.13, b.13]),
			and(in: [a.14, b.14]),
			and(in: [a.15, b.15])
		]
	}
	`
	parser := NewParser(Lexer{
		Source: source,
	})
	definitions, err := parser.Parse()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	support := make(map[string]ChipDefinition)
	for _, d := range definitions {
		support[d.Name] = d
	}
	compiler := NewCompiler(support)
	and, err := compiler.Compile(support["and_16"])
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	tests := []struct {
		a      []byte
		b      []byte
		expect []byte
	}{
		{
			a:      []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			b:      []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expect: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			a:      []byte{1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0},
			b:      []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
			expect: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
		},
		{
			a:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			b:      []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
			expect: []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
		},
		{
			a:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			b:      []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expect: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			a:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			b:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			expect: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v & %v", test.a, test.b), func(t *testing.T) {
			err := compiler.breadboard.SetGroup(and.Inputs["a"], test.a)
			if err != nil {
				t.Errorf("unexpected error when setting a: %s", err)
			}
			err = compiler.breadboard.SetGroup(and.Inputs["b"], test.b)
			if err != nil {
				t.Errorf("unexpected error when setting b: %s", err)
			}

			output, err := compiler.breadboard.GetGroup(and.Outputs[0])
			if err != nil {
				t.Errorf("unexpected error when extracting output: %s", err)
			}
			if !reflect.DeepEqual(output, test.expect) {
				t.Errorf("expected %v but got %v", test.expect, output)
			}
		})
	}
}

func TestAnd16To1(t *testing.T) {
	source := `
	chip not (in: 1) -> (1) {
		out nand(in: [in.0, 1])
	}
	
	chip and (in: 2) -> (1) {
		out not(in: nand(in: [in.0, in.1]))
	}

	chip and_16_to_1 (a: 16, b: 1) -> (16) {
		out [
			and(in: [a.0, b]),
			and(in: [a.1, b]),
			and(in: [a.2, b]),
			and(in: [a.3, b]),
			and(in: [a.4, b]),
			and(in: [a.5, b]),
			and(in: [a.6, b]),
			and(in: [a.7, b]),
			and(in: [a.8, b]),
			and(in: [a.9, b]),
			and(in: [a.10, b]),
			and(in: [a.11, b]),
			and(in: [a.12, b]),
			and(in: [a.13, b]),
			and(in: [a.14, b]),
			and(in: [a.15, b])
		]
	}
	`
	parser := NewParser(Lexer{
		Source: source,
	})
	definitions, err := parser.Parse()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	support := make(map[string]ChipDefinition)
	for _, d := range definitions {
		support[d.Name] = d
	}
	compiler := NewCompiler(support)
	and, err := compiler.Compile(support["and_16_to_1"])
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	tests := []struct {
		a      []byte
		b      byte
		expect []byte
	}{
		{
			a:      []byte{1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1},
			b:      1,
			expect: []byte{1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1},
		},
		{
			a:      []byte{1, 0, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 0},
			b:      0,
			expect: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v & %v", test.a, test.b), func(t *testing.T) {
			err := compiler.breadboard.SetGroup(and.Inputs["a"], test.a)
			if err != nil {
				t.Errorf("unexpected error when setting a: %s", err)
			}
			compiler.breadboard.Set(Pin{ID: and.Inputs["b"], Index: 0}, test.b)

			output, err := compiler.breadboard.GetGroup(and.Outputs[0])
			if err != nil {
				t.Errorf("unexpected error when extracting output: %s", err)
			}
			if !reflect.DeepEqual(output, test.expect) {
				t.Errorf("expected %v but got %v", test.expect, output)
			}
		})
	}
}

func TestMux2(t *testing.T) {
	source := `
	chip not (in: 1) -> (1) {
		out nand(in: [in.0, 1])
	}
	
	chip and (in: 2) -> (1) {
		out not(in: nand(in: [in.0, in.1]))
	}

	chip and_16_to_1 (a: 16, b: 1) -> (16) {
		out [
			and(in: [a.0, b]),
			and(in: [a.1, b]),
			and(in: [a.2, b]),
			and(in: [a.3, b]),
			and(in: [a.4, b]),
			and(in: [a.5, b]),
			and(in: [a.6, b]),
			and(in: [a.7, b]),
			and(in: [a.8, b]),
			and(in: [a.9, b]),
			and(in: [a.10, b]),
			and(in: [a.11, b]),
			and(in: [a.12, b]),
			and(in: [a.13, b]),
			and(in: [a.14, b]),
			and(in: [a.15, b])
		]
	}
	
	chip or (in: 2) -> (1) {
		out nand(in: [not(in: in.0), not(in: in.1)])
	}
	
	chip or_16 (a: 16, b: 16) -> (16) {
		out [
			or(in: [a.0, b.0]),
			or(in: [a.1, b.1]),
			or(in: [a.2, b.2]),
			or(in: [a.3, b.3]),
			or(in: [a.4, b.4]),
			or(in: [a.5, b.5]),
			or(in: [a.6, b.6]),
			or(in: [a.7, b.7]),
			or(in: [a.8, b.8]),
			or(in: [a.9, b.9]),
			or(in: [a.10, b.10]),
			or(in: [a.11, b.11]),
			or(in: [a.12, b.12]),
			or(in: [a.13, b.13]),
			or(in: [a.14, b.14]),
			or(in: [a.15, b.15])
		]
	}

	chip mux_2 (s: 1, a: 16, b: 16) -> (16) {
		out or_16(
			a: and_16_to_1(a: a, b: not(in: s)),
			b: and_16_to_1(a: b, b: s)
		)
	}
	`
	parser := NewParser(Lexer{
		Source: source,
	})
	definitions, err := parser.Parse()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	support := make(map[string]ChipDefinition)
	for _, d := range definitions {
		support[d.Name] = d
	}
	compiler := NewCompiler(support)
	and, err := compiler.Compile(support["mux_2"])
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	tests := []struct {
		s      byte
		a      []byte
		b      []byte
		expect []byte
	}{
		{
			s:      0,
			a:      []byte{1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0},
			b:      []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
			expect: []byte{1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0},
		},
		{
			s:      1,
			a:      []byte{1, 0, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 0},
			b:      []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
			expect: []byte{0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v & %v", test.a, test.b), func(t *testing.T) {
			compiler.breadboard.Set(Pin{ID: and.Inputs["s"], Index: 0}, test.s)
			err := compiler.breadboard.SetGroup(and.Inputs["a"], test.a)
			if err != nil {
				t.Errorf("unexpected error when setting a: %s", err)
			}
			err = compiler.breadboard.SetGroup(and.Inputs["b"], test.b)
			if err != nil {
				t.Errorf("unexpected error when setting b: %s", err)
			}

			output, err := compiler.breadboard.GetGroup(and.Outputs[0])
			if err != nil {
				t.Errorf("unexpected error when extracting output: %s", err)
			}
			if !reflect.DeepEqual(output, test.expect) {
				t.Errorf("expected %v but got %v", test.expect, output)
			}
		})
	}
}
