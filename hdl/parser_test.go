package hdl

import (
	"errors"
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		src  string
		stmt Statement
		err  error
	}{
		{
			src: `chip and (a: 1, b: 1) -> (1) {}`,
			stmt: ChipStatement{
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
			stmt: ChipStatement{
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
			src: `use "../mux/mux.hdl"`,
			stmt: UseStatement{
				FileName: "../mux/mux.hdl",
			},
			err: nil,
		},
		{
			src: `chip not16 (n: 16) -> (16) {}`,
			stmt: ChipStatement{
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
			stmt not16 (n: 16) -> (1, 1) {
				out n
				out 1
			}`,
			stmt: ChipStatement{
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
						Expression: IdentifierExpression{Identifier: "n"},
					},
					OutStatement{
						Expression: IntegerExpression{Integer: 1},
					},
				},
			},
			err: nil,
		},
		{
			src: `
			stmt and (a: 1, b: 1) -> (1) {
				out nand(a: not(a: a.0), b: not(a: b.0))
			}`,
			stmt: ChipStatement{
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
						Expression: CallExpression{
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
			stmt not (in: 1) -> (1) {
				out nand(in: [in.0, 1])
			}`,
			stmt: ChipStatement{
				Name: "not",
				Inputs: map[string]byte{
					"in": 1,
				},
				Outputs: []byte{
					1,
				},
				Body: []Statement{
					OutStatement{
						Expression: CallExpression{
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
		{
			src: `
			stmt flerp (in: 16) -> (16) {
				set a, b, c, d = dmux_4(s: [0, 0], in: in)
				out a
			}`,
			stmt: ChipStatement{
				Name: "flerp",
				Inputs: map[string]byte{
					"in": 16,
				},
				Outputs: []byte{
					16,
				},
				Body: []Statement{
					SetStatement{
						Identifiers: []string{"a", "b", "c", "d"},
						Expression: CallExpression{
							Name: "dmux_4",
							Args: map[string]Expression{
								"s": ArrayExpression{
									Values: []Expression{
										IntegerExpression{Integer: 0},
										IntegerExpression{Integer: 0},
									},
								},
								"in": IdentifierExpression{Identifier: "in"},
							},
						},
					},
					OutStatement{Expression: IdentifierExpression{Identifier: "a"}},
				},
			},
			err: nil,
		},
		{
			src: `
			stmt test (in: 1) -> (1, 1) {
				set regular = in
				out regular
			}`,
			stmt: ChipStatement{
				Name: "test",
				Inputs: map[string]byte{
					"in": 1,
				},
				Outputs: []byte{
					1,
					1,
				},
				Body: []Statement{
					SetStatement{
						Identifiers: []string{"regular"},
						Expression: IdentifierExpression{
							Identifier: "in",
						},
					},
					OutStatement{Expression: IdentifierExpression{Identifier: "regular"}},
				},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			parser := Parser{lexer: LoadedLexer(test.src)}
			ch, err := parser.Parse()
			if !errors.Is(err, test.err) {
				t.Errorf("expected err to be %v but got %v", test.err, err)
			}
			if !reflect.DeepEqual(ch[0], test.stmt) {
				t.Errorf("expected stmt to equal %v but got %v", test.stmt, ch[0])
			}
		})
	}
}
