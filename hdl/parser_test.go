package hdl

import (
	"errors"
	"reflect"
	"testing"
)

func TestChipParser_ParseChip(t *testing.T) {
	tests := []struct {
		src  string
		chip Chip
		err  error
	}{
		{
			src: `chip and (a: 1, b: 1) -> (1) {}`,
			chip: Chip{
				name: "and",
				inputs: map[string]int{
					"a": 1,
					"b": 1,
				},
				outputs: []int{
					1,
				},
				body: []Statement{},
			},
			err: nil,
		},
		{
			src: `chip mux (s: 2, n: 16) -> (16, 16, 16, 16) {}`,
			chip: Chip{
				name: "mux",
				inputs: map[string]int{
					"s": 2,
					"n": 16,
				},
				outputs: []int{
					16,
					16,
					16,
					16,
				},
				body: []Statement{},
			},
			err: nil,
		},
		{
			src: `chip not16 (n: 16) -> (16) {}`,
			chip: Chip{
				name: "not16",
				inputs: map[string]int{
					"n": 16,
				},
				outputs: []int{
					16,
				},
				body: []Statement{},
			},
			err: nil,
		},
		{
			src: `
			chip not16 (n: 16) -> (16) {
				set one = 1
			}`,
			chip: Chip{
				name: "not16",
				inputs: map[string]int{
					"n": 16,
				},
				outputs: []int{
					16,
				},
				body: []Statement{
					SetStatement{
						identifier: "one",
						expression: IntegerExpression{literal: 1},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			chip not16 (n: 16) -> (1, 1) {
				set one = 1
				out n
				out 1
			}`,
			chip: Chip{
				name: "not16",
				inputs: map[string]int{
					"n": 16,
				},
				outputs: []int{
					1,
					1,
				},
				body: []Statement{
					SetStatement{
						identifier: "one",
						expression: IntegerExpression{literal: 1},
					},
					OutStatement{
						expression: IdentifierExpression{literal: "n"},
					},
					OutStatement{
						expression: IntegerExpression{literal: 1},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			chip not (n: 1) -> (1, 1) {
				out nand(a: n, b: 0)
				out n
			}`,
			chip: Chip{
				name: "not",
				inputs: map[string]int{
					"n": 1,
				},
				outputs: []int{
					1,
					1,
				},
				body: []Statement{
					OutStatement{
						expression: CallExpression{
							name: "nand",
							args: map[string]Expression{
								"a": IdentifierExpression{literal: "n"},
								"b": IntegerExpression{literal: 0},
							},
						},
					},
					OutStatement{
						expression: IdentifierExpression{literal: "n"},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			chip and (a: 1, b: 1) -> (1) {
				out nand(a: not(a: a), b: not(a: b))
			}`,
			chip: Chip{
				name: "and",
				inputs: map[string]int{
					"a": 1,
					"b": 1,
				},
				outputs: []int{
					1,
				},
				body: []Statement{
					OutStatement{
						expression: CallExpression{
							name: "nand",
							args: map[string]Expression{
								"a": CallExpression{
									name: "not",
									args: map[string]Expression{
										"a": IdentifierExpression{literal: "a"},
									},
								},
								"b": CallExpression{
									name: "not",
									args: map[string]Expression{
										"a": IdentifierExpression{literal: "b"},
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
			parser := Parser{lexer: lexer{src: test.src}}
			ch, err := parser.Parse()
			if !errors.Is(err, test.err) {
				t.Errorf("expected err to be %v but got %v", test.err, err)
			}
			if !reflect.DeepEqual(ch, test.chip) {
				t.Errorf("expected chip to equal %v but got %v", test.chip, ch)
			}
		})
	}
}
