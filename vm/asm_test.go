package vm

import (
	"log"
	"testing"
)

func TestPush_Get(t *testing.T) {
	asm, err := Push{}.Get()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",
	}
	for i := range asm {
		actual := asm[i].Get()
		if expected[i] != actual {
			t.Errorf("expected %s but got %s", expected[i], actual)
		}
	}
}

func TestPop_Get(t *testing.T) {
	tests := []struct {
		name     string
		targets  Targets
		expected []string
	}{
		{
			name:    "pop to D",
			targets: Target(D{}),
			expected: []string{
				"@SP",
				"AM=M-1",
				"D=M",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Pop{}.Get()
			if err != nil {
				log.Fatal(err)
			}
			for i := range actual {
				if actual[i].Get() != test.expected[i] {
					t.Errorf("expected %s but got %s", test.expected[i], actual[i])
				}
			}
		})
	}
	asm, err := Pop{}.Get()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"@SP",
		"AM=M-1",
		"D=M",
	}
	for i := range asm {
		actual := asm[i].Get()
		if expected[i] != actual {
			t.Errorf("expected %s but got %s", expected[i], actual)
		}
	}
}
