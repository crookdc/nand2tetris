package hdl

import (
	"testing"
)

func TestLexer_Next(t *testing.T) {
	var tests = []struct {
		name   string
		src    string
		tokens []token
	}{
		{
			name: "source code for chip implementation of NOT gate",
			src: `
			chip not (n: 1) -> (1) {
				out {
					nand(n=[n, 1])
				}
			}
			`,
			tokens: []token{
				{
					variant: chip,
					literal: "chip",
				},
				{
					variant: identifier,
					literal: "not",
				},
				{
					variant: leftParenthesis,
					literal: "(",
				},
				{
					variant: identifier,
					literal: "n",
				},
				{
					variant: colon,
					literal: ":",
				},
				{
					variant: integer,
					literal: "1",
				},
				{
					variant: rightParenthesis,
					literal: ")",
				},
				{
					variant: arrow,
					literal: "->",
				},
				{
					variant: leftParenthesis,
					literal: "(",
				},
				{
					variant: integer,
					literal: "1",
				},
				{
					variant: rightParenthesis,
					literal: ")",
				},
				{
					variant: leftCurlyBrace,
					literal: "{",
				},
				{
					variant: out,
					literal: "out",
				},
				{
					variant: leftCurlyBrace,
					literal: "{",
				},
				{
					variant: identifier,
					literal: "nand",
				},
				{
					variant: leftParenthesis,
					literal: "(",
				},
				{
					variant: identifier,
					literal: "n",
				},
				{
					variant: equals,
					literal: "=",
				},
				{
					variant: leftBracket,
					literal: "[",
				},
				{
					variant: identifier,
					literal: "n",
				},
				{
					variant: comma,
					literal: ",",
				},
				{
					variant: integer,
					literal: "1",
				},
				{
					variant: rightBracket,
					literal: "]",
				},
				{
					variant: rightParenthesis,
					literal: ")",
				},
				{
					variant: rightCurlyBrace,
					literal: "}",
				},
				{
					variant: rightCurlyBrace,
					literal: "}",
				},
				{
					variant: eof,
					literal: "",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer{src: test.src}
			i := 0
			for {
				tok, err := l.next()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				expected := test.tokens[i]
				if tok.variant != expected.variant {
					t.Errorf("expected variant '%v' but got '%v'", expected.variant, tok.variant)
				}
				if tok.literal != expected.literal {
					t.Errorf("expected Literal '%v' but got '%v'", expected.literal, tok.literal)
				}
				i++
				if tok.variant == eof {
					break
				}
			}
			if i != len(test.tokens) {
				t.Errorf("expected %d tokens but got %d", len(test.tokens), i)
			}
		})
	}
}
