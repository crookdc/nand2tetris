package simulator

import (
	"fmt"
	"testing"
)

func TestHigh(t *testing.T) {
	tests := []struct {
		word     uint16
		n        int
		expected bool
	}{
		{
			word:     0b0000_0000_0000_0000,
			n:        16,
			expected: false,
		},
		{
			word:     0b1000_0000_0000_0000,
			n:        16,
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%b << %d", test.word, test.n), func(t *testing.T) {
			actual := high(test.word, test.n)
			if test.expected != actual {
				t.Errorf("expected %v but got %v", test.expected, actual)
			}
		})
	}
}
