package asm

import (
	"github.com/crookdc/nand2tetris/lexer"
	"reflect"
	"testing"
)

func TestParser_next(t *testing.T) {
	var tests = []struct {
		src string
		res []instruction
	}{
		{
			src: "@17\n",
			res: []instruction{
				load{value: lexer.Token[variant]{
					Variant: integer,
					Literal: "17",
				}},
			},
		},
		{
			src: "A=D+1\n",
			res: []instruction{
				compute{
					dest: &lexer.Token[variant]{
						Variant: identifier,
						Literal: "A",
					},
					comp: "D+1",
					jump: nil,
				},
			},
		},
		{
			src: "A;JGT\n",
			res: []instruction{
				compute{
					dest: nil,
					comp: "A",
					jump: &lexer.Token[variant]{
						Variant: jgt,
						Literal: "JGT",
					},
				},
			},
		},
		{
			src: "@i\nD=A\nD=D+1;JNE\n",
			res: []instruction{
				load{
					value: lexer.Token[variant]{
						Variant: identifier,
						Literal: "i",
					},
				},
				compute{
					dest: &lexer.Token[variant]{
						Variant: identifier,
						Literal: "D",
					},
					comp: "A",
					jump: nil,
				},
				compute{
					dest: &lexer.Token[variant]{
						Variant: identifier,
						Literal: "D",
					},
					comp: "D+1",
					jump: &lexer.Token[variant]{
						Variant: jne,
						Literal: "JNE",
					},
				},
			},
		},
		{
			src: "(loop)\n@1234\nD=A+1\n@loop\n0;JMP\n",
			res: []instruction{
				label{
					value: lexer.Token[variant]{
						Variant: identifier,
						Literal: "loop",
					},
				},
				load{
					value: lexer.Token[variant]{
						Variant: integer,
						Literal: "1234",
					},
				},
				compute{
					dest: &lexer.Token[variant]{
						Variant: identifier,
						Literal: "D",
					},
					comp: "A+1",
					jump: nil,
				},
				load{
					value: lexer.Token[variant]{
						Variant: identifier,
						Literal: "loop",
					},
				},
				compute{
					dest: nil,
					comp: "0",
					jump: &lexer.Token[variant]{
						Variant: jmp,
						Literal: "JMP",
					},
				},
			},
		},
		{
			src: "// This is just a friendly comment\n(loop)\n@1234\n",
			res: []instruction{
				label{
					value: lexer.Token[variant]{
						Variant: identifier,
						Literal: "loop",
					},
				},
				load{
					value: lexer.Token[variant]{
						Variant: integer,
						Literal: "1234",
					},
				},
			},
		},
	}
	for _, test := range tests {
		ps := parser{lexer: LoadedLexer(test.src)}
		ins, _ := ps.next()
		for i := 0; ins != nil; i++ {
			if !reflect.DeepEqual(test.res[i], ins) {
				t.Errorf("expected %+v but got %+v", test.res[i].Literal(), ins.Literal())
			}
			ins, _ = ps.next()
		}
	}
}
