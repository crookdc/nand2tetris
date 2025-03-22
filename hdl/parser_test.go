package hdl

import (
	"errors"
	"reflect"
	"testing"
)

func TestChipParser_ParseChip(t *testing.T) {
	tests := []struct {
		src  string
		chip ChipDefinition
		err  error
	}{
		{
			src: `chip and (a: 1, b: 1) -> (1) {}`,
			chip: ChipDefinition{
				Name: "and",
				Inputs: map[string]byte{
					"a": 1,
					"b": 1,
				},
				Outputs: []byte{
					1,
				},
				Body: []Statement{},
			},
			err: nil,
		},
		{
			src: `chip mux (s: 2, n: 16) -> (16, 16, 16, 16) {}`,
			chip: ChipDefinition{
				Name: "mux",
				Inputs: map[string]byte{
					"s": 2,
					"n": 16,
				},
				Outputs: []byte{
					16,
					16,
					16,
					16,
				},
				Body: []Statement{},
			},
			err: nil,
		},
		{
			src: `chip not16 (n: 16) -> (16) {}`,
			chip: ChipDefinition{
				Name: "not16",
				Inputs: map[string]byte{
					"n": 16,
				},
				Outputs: []byte{
					16,
				},
				Body: []Statement{},
			},
			err: nil,
		},
		{
			src: `
			chip not16 (n: 16) -> (1, 1) {
				out n
				out 1
			}`,
			chip: ChipDefinition{
				Name: "not16",
				Inputs: map[string]byte{
					"n": 16,
				},
				Outputs: []byte{
					1,
					1,
				},
				Body: []Statement{
					OutStatement{
						expression: IdentifierExpression{Identifier: "n"},
					},
					OutStatement{
						expression: IntegerExpression{Integer: 1},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			chip and (a: 1, b: 1) -> (1) {
				out nand(a: not(a: a.0), b: not(a: b.0))
			}`,
			chip: ChipDefinition{
				Name: "and",
				Inputs: map[string]byte{
					"a": 1,
					"b": 1,
				},
				Outputs: []byte{
					1,
				},
				Body: []Statement{
					OutStatement{
						expression: CallExpression{
							Name: "nand",
							Args: map[string]Expression{
								"a": CallExpression{
									Name: "not",
									Args: map[string]Expression{
										"a": IndexedExpression{Identifier: "a", Index: 0},
									},
								},
								"b": CallExpression{
									Name: "not",
									Args: map[string]Expression{
										"a": IndexedExpression{Identifier: "b", Index: 0},
									},
								},
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			chip not (in: 1) -> (1) {
				out nand(in: [in.0, 1])
			}`,
			chip: ChipDefinition{
				Name: "not",
				Inputs: map[string]byte{
					"in": 1,
				},
				Outputs: []byte{
					1,
				},
				Body: []Statement{
					OutStatement{
						expression: CallExpression{
							Name: "nand",
							Args: map[string]Expression{
								"in": ArrayExpression{
									Values: []Expression{
										IndexedExpression{
											Identifier: "in",
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
			err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			parser := Parser{lexer: Lexer{Source: test.src}}
			ch, err := parser.Parse()
			if !errors.Is(err, test.err) {
				t.Errorf("expected err to be %v but got %v", test.err, err)
			}
			if !reflect.DeepEqual(ch[0], test.chip) {
				t.Errorf("expected chip to equal %v but got %v", test.chip, ch[0])
			}
		})
	}
}
