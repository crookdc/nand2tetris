package asm

import "testing"

func TestWrap(t *testing.T) {
	tests := []struct {
		n int
		r [16]byte
	}{
		{
			n: 60432,
			r: [16]byte{1, 1, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
		},
	}
	for _, test := range tests {
		r := wrap(test.n)
		if r != test.r {
			t.Errorf("expected %v but got %v", test.r, r)
		}
	}
}
