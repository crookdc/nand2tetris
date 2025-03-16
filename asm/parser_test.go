package asm

import (
	"reflect"
	"testing"
)

func TestParser_next(t *testing.T) {
	var assertions = []struct {
		src string
		res []instruction
	}{
		{
			src: "@17\n",
			res: []instruction{
				load{value: token{
					variant: integer,
					literal: "17",
				}},
			},
		},
		{
			src: "A=D+1\n",
			res: []instruction{
				compute{
					dest: &token{
						variant: identifier,
						literal: "A",
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
					jump: &token{
						variant: jgt,
						literal: "JGT",
					},
				},
			},
		},
		{
			src: "@i\nD=A\nD=D+1;JNE\n",
			res: []instruction{
				load{
					value: token{
						variant: identifier,
						literal: "i",
					},
				},
				compute{
					dest: &token{
						variant: identifier,
						literal: "D",
					},
					comp: "A",
					jump: nil,
				},
				compute{
					dest: &token{
						variant: identifier,
						literal: "D",
					},
					comp: "D+1",
					jump: &token{
						variant: jne,
						literal: "JNE",
					},
				},
			},
		},
		{
			src: "(loop)\n@1234\nD=A+1\n@loop\n0;JMP\n",
			res: []instruction{
				label{
					value: token{
						variant: identifier,
						literal: "loop",
					},
				},
				load{
					value: token{
						variant: integer,
						literal: "1234",
					},
				},
				compute{
					dest: &token{
						variant: identifier,
						literal: "D",
					},
					comp: "A+1",
					jump: nil,
				},
				load{
					value: token{
						variant: identifier,
						literal: "loop",
					},
				},
				compute{
					dest: nil,
					comp: "0",
					jump: &token{
						variant: jmp,
						literal: "JMP",
					},
				},
			},
		},
	}
	for _, assert := range assertions {
		ps := parser{lexer: lexer{src: assert.src}}
		ins, _ := ps.next()
		for i := 0; ins != nil; i++ {
			if !reflect.DeepEqual(assert.res[i], ins) {
				t.Errorf("expected %+v but got %+v", assert.res[i], ins)
			}
			ins, _ = ps.next()
		}
	}
}
