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
