package hdl

import "testing"

func TestBreadboard_Connect(t *testing.T) {
	breadboard := NewBreadboard()
	i := breadboard.Allocate(1, nil)
	j := breadboard.Allocate(1, nil)
	breadboard.Connect(
		Wire{
			Head: Pin{
				ID:    i,
				Index: 0,
			},
			Tail: Pin{
				ID:    j,
				Index: 0,
			},
		},
	)
	breadboard.Set(Pin{ID: i, Index: 0}, 1)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: j, Index: 0}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	k := breadboard.Allocate(2, nil)
	breadboard.Connect(Wire{
		Head: Pin{
			ID:    j,
			Index: 0,
		},
		Tail: Pin{
			ID:    k,
			Index: 1,
		},
	})
	Tick(breadboard)
	
	if breadboard.Get(Pin{ID: j, Index: 0}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: k, Index: 1}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}

	breadboard.Set(Pin{ID: i, Index: 0}, 0)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: i, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: j, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: k, Index: 1}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}

	breadboard.Set(Pin{ID: j, Index: 0}, 1)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: i, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: j, Index: 0}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: k, Index: 1}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}
}

func TestBreadboard_ConnectGroup(t *testing.T) {
	breadboard := NewBreadboard()
	i := breadboard.Allocate(8, nil)
	j := breadboard.Allocate(8, nil)
	if err := breadboard.ConnectGroup(i, j); err != nil {
		t.Errorf("unexpected error %v", err)
	}
	breadboard.Set(Pin{ID: i, Index: 0}, 1)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: j, Index: 0}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	breadboard.Set(Pin{ID: i, Index: 5}, 1)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: j, Index: 5}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	k := breadboard.Allocate(8, nil)
	if err := breadboard.ConnectGroup(j, k); err != nil {
		t.Errorf("unexpected error %v", err)
	}
	Tick(breadboard)
	if breadboard.Get(Pin{ID: k, Index: 0}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}
	if breadboard.Get(Pin{ID: k, Index: 5}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}

	breadboard.Set(Pin{ID: i, Index: 0}, 0)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: i, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: j, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: k, Index: 0}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}

	breadboard.Set(Pin{ID: j, Index: 4}, 1)
	Tick(breadboard)

	if breadboard.Get(Pin{ID: i, Index: 4}) != 0 {
		t.Errorf("expected 0 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: j, Index: 4}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: j, Index: 0}))
	}
	if breadboard.Get(Pin{ID: k, Index: 4}) != 1 {
		t.Errorf("expected 1 but got %d", breadboard.Get(Pin{ID: k, Index: 1}))
	}
}
