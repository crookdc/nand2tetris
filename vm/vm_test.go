package vm

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		src      string
		expected []string
		ok       bool
	}{
		{
			src: "push constant 0",
			expected: []string{
				"@0",
				"D=A",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push constant 24565",
			expected: []string{
				"@24565",
				"D=A",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push constant -256",
			ok:  false,
		},
		{
			src: "push constant 32769",
			ok:  false,
		},
		{
			src: "push local 17",
			expected: []string{
				"@LCL",
				"D=M",
				"@17",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push local 0",
			expected: []string{
				"@LCL",
				"D=M",
				"@0",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push local -1",
			ok:  false,
		},
		{
			src: "push arg 110",
			expected: []string{
				"@ARG",
				"D=M",
				"@110",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push arg 0",
			expected: []string{
				"@ARG",
				"D=M",
				"@0",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push arg -1",
			ok:  false,
		},
		{
			src: "push this 110",
			expected: []string{
				"@THIS",
				"D=M",
				"@110",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push this 0",
			expected: []string{
				"@THIS",
				"D=M",
				"@0",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push this -1",
			ok:  false,
		},
		{
			src: "push that 110",
			expected: []string{
				"@THAT",
				"D=M",
				"@110",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push that 0",
			expected: []string{
				"@THAT",
				"D=M",
				"@0",
				"A=D+A",
				"D=M",
				"@SP",
				"A=M",
				"M=D",
				"@SP",
				"M=M+1",
			},
			ok: true,
		},
		{
			src: "push that -1",
			ok:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			actual, err := Evaluate(strings.NewReader(test.src))
			if (err == nil && !test.ok) || (err != nil && test.ok) {
				t.Errorf("expected non-nil error but got %v", err)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %v but got %v", test.expected, actual)
			}
		})
	}
}
