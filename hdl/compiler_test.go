package hdl

import (
	"testing"
)

func TestNAND(t *testing.T) {
	breadboard := NewBreadboard()
	input, output := NAND(breadboard)
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 1,
	}, 1)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 0 {
		t.Errorf("expected output to be 0 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
	breadboard.Set(Pin{
		ID:    input,
		Index: 0,
	}, 0)
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output, Index: 0}) != 1 {
		t.Errorf("expected output to be 1 but got %v", breadboard.Get(Pin{ID: output, Index: 0}))
	}
}

func TestDFF(t *testing.T) {
	breadboard := NewBreadboard()
	input, output := DFF(breadboard)
	breadboard.Set(Pin{ID: input}, 1)
	if breadboard.Get(Pin{ID: output}) != 0 {
		t.Errorf("expected DFF output to be 0 before clock")
	}
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output}) != 1 {
		t.Errorf("expected DFF output to be 1 after clock")
	}
	breadboard.Set(Pin{ID: input}, 0)
	if breadboard.Get(Pin{ID: output}) != 1 {
		t.Errorf("expected DFF output to be 1 before clock")
	}
	Tick(breadboard)
	if breadboard.Get(Pin{ID: output}) != 0 {
		t.Errorf("expected DFF output to be 0 after clock")
	}
}
