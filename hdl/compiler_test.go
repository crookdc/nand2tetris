package hdl

import "testing"

func TestNAND_Signal(t *testing.T) {
	nand := &NAND{
		a:      &Pin{},
		b:      &Pin{},
		output: &Pin{},
	}
	nand.a.Binding = nand
	nand.a.Set(1)
	nand.b.Binding = nand
	nand.b.Set(1)

	not := &NAND{
		a:      &Pin{},
		b:      &Pin{signal: 1},
		output: &Pin{},
	}
	not.a.Binding = not
	not.b.Binding = not
	and := &NAND{
		a:      &Pin{},
		b:      &Pin{},
		output: not.a,
	}
	and.a.Binding = and
	and.a.Set(1)
	and.b.Binding = and
	and.b.Set(1)
	if not.output.signal != 1 {
		panic("")
	}

	and.b.Set(0)
	if not.output.signal != 0 {
		panic("")
	}
}
