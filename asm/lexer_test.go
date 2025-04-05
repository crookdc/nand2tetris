package asm

import (
	"fmt"
	"testing"
)

func TestLexer_next(t *testing.T) {
	var assertions = []struct {
		src    string
		tokens []token
	}{
		{
			src: "@17\n",
			tokens: []token{
				{
					variant: at,
					literal: "@",
				},
				{
					variant: integer,
					literal: "17",
				},
				{
					variant: linefeed,
					literal: "\n",
				},
			},
		},
		{
			src: "ADM=A+1\n",
			tokens: []token{
				{
					variant: identifier,
					literal: "ADM",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: identifier,
					literal: "A",
				},
				{
					variant: plus,
					literal: "+",
				},
				{
					variant: integer,
					literal: "1",
				},
				{
					variant: linefeed,
					literal: "\n",
				},
			},
		},
		{
			src: "//this is a comment\n@2000",
			tokens: []token{
				{
					variant: at,
					literal: "@",
				},
				{
					variant: integer,
					literal: "2000",
				},
			},
		},
		{
			src: "A;D;M",
			tokens: []token{
				{
					variant: identifier,
					literal: "A",
				},
				{
					variant: semicolon,
					literal: ";",
				},
				{
					variant: identifier,
					literal: "D",
				},
				{
					variant: semicolon,
					literal: ";",
				},
				{
					variant: identifier,
					literal: "M",
				},
			},
		},
		{
			src: "D;JGT",
			tokens: []token{
				{
					variant: identifier,
					literal: "D",
				},
				{
					variant: semicolon,
					literal: ";",
				},
				{
					variant: jgt,
					literal: "JGT",
				},
			},
		},
		{
			src: "M=-1",
			tokens: []token{
				{
					variant: identifier,
					literal: "M",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: minus,
					literal: "-",
				},
				{
					variant: integer,
					literal: "1",
				},
			},
		},
		{
			src: "M=!M",
			tokens: []token{
				{
					variant: identifier,
					literal: "M",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: bang,
					literal: "!",
				},
				{
					variant: identifier,
					literal: "M",
				},
			},
		},
		{
			src: "M=D",
			tokens: []token{
				{
					variant: identifier,
					literal: "M",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: identifier,
					literal: "D",
				},
			},
		},
		{
			src: "@10\nD=A+1\nM=D\nD;JNE",
			tokens: []token{
				{
					variant: at,
					literal: "@",
				},
				{
					variant: integer,
					literal: "10",
				},
				{
					variant: linefeed,
					literal: "\n",
				},
				{
					variant: identifier,
					literal: "D",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: identifier,
					literal: "A",
				},
				{
					variant: plus,
					literal: "+",
				},
				{
					variant: integer,
					literal: "1",
				},
				{
					variant: linefeed,
					literal: "\n",
				},
				{
					variant: identifier,
					literal: "M",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: identifier,
					literal: "D",
				},
				{
					variant: linefeed,
					literal: "\n",
				},
				{
					variant: identifier,
					literal: "D",
				},
				{
					variant: semicolon,
					literal: ";",
				},
				{
					variant: jne,
					literal: "JNE",
				},
			},
		},
	}
	for _, a := range assertions {
		t.Run(fmt.Sprintf("given src %v", a.src), func(t *testing.T) {
			l := lexer{src: a.src}
			i := 0
			var err error
			var tok token
			for tok, err = l.next(); tok.variant != eof && err == nil; {
				if a.tokens[i].variant != tok.variant {
					t.Errorf("expected tok variant %v but got %v", a.tokens[i].variant, tok.variant)
				}
				if a.tokens[i].literal != tok.literal {
					t.Errorf("expected tok variant %v but got %v", a.tokens[i].variant, tok.variant)
				}
				tok, err = l.next()
				i += 1
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
