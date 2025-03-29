package hdl

import (
	"errors"
)

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidIndex     = errors.New("index out of range")
	ErrNonUniformGroups = errors.New("groups are not uniform")
)

type signal struct {
	value byte
}

func (s *signal) set(value byte) {
	if value != 1 && value != 0 {
		panic("tried to set invalid signal value")
	}
	s.value = value
}

type Pin struct {
	ID    ID
	Index int
}

type Wire struct {
	Head Pin
	Tail Pin
}

type Callback func(ID, []byte)

type ID = int

func NewBreadboard() *Breadboard {
	breadboard := &Breadboard{
		groups: make([]group, 0),
		wires:  make(map[Pin][]Pin),
		changeset: &changeset{
			keys:  make(map[Pin]struct{}),
			queue: make([]Pin, 0),
		},
	}
	breadboard.CLK = breadboard.Allocate(1, nil)
	breadboard.Zero = breadboard.Allocate(1, nil)
	breadboard.One = breadboard.Allocate(1, nil)
	breadboard.Set(Pin{ID: breadboard.One, Index: 0}, 1)
	return breadboard
}

type group struct {
	callback Callback
	pins     []signal
}

type changeset struct {
	keys  map[Pin]struct{}
	queue []Pin
}

func (cs *changeset) enqueue(v Pin) {
	if _, ok := cs.keys[v]; ok {
		return
	}
	cs.queue = append(cs.queue, v)
	cs.keys[v] = struct{}{}
}

func (cs *changeset) dequeue() Pin {
	if len(cs.queue) == 0 {
		panic("dequeue called on empty queue")
	}
	element := cs.queue[0]
	cs.queue = cs.queue[1:]
	delete(cs.keys, element)
	return element
}

func (cs *changeset) more() bool {
	return len(cs.queue) > 0
}

type Breadboard struct {
	CLK       ID
	Zero      ID
	One       ID
	groups    []group
	wires     map[Pin][]Pin
	changeset *changeset
}

// SizeOf returns the length of the group registered under the provided ID. An error is returned if the ID is not
// registered on the Breadboard or is otherwise invalid.
func (b *Breadboard) SizeOf(id ID) (int, error) {
	if !b.exists(id) {
		return 0, ErrInvalidID
	}
	return len(b.groups[id].pins), nil
}

// Allocate allocates a new pin group of the provided size together with a callback to be called when a new value is set
// on any of the groups within the group. The returned ID is the ID to be used when retrieving or setting values within
// the group. See [hdl.Breadboard.Set], [hdl.Breadboard.Get], [hdl.Breadboard.Get] and [hdl.Breadboard.GetGroup].
func (b *Breadboard) Allocate(count int, cb Callback) ID {
	id := len(b.groups)
	b.groups = append(b.groups, group{
		callback: cb,
		pins:     make([]signal, count),
	})
	return id
}

// Connect causes one pin (the tail) to always follow the value of the other (the head). The connection is one-way such
// that whenever the head pin changes the tail pin will take on the same value but if the tail pin changes on its own
// accord then the head pin will keep its value.
func (b *Breadboard) Connect(wire Wire) {
	if err := b.validate(wire.Head); err != nil {
		panic(err)
	}
	if err := b.validate(wire.Tail); err != nil {
		panic(err)
	}
	b.connect(wire)
}

func (b *Breadboard) connect(wire Wire) {
	b.wires[wire.Head] = append(b.wires[wire.Head], wire.Tail)
	b.changeset.enqueue(wire.Head)
}

// ConnectGroup is a convenience method that connects all groups of the groups identified by the supplied ID's with head
// as the driver. It is negligibly quicker than calling Connect several times while iterating over a known number of
// groups since it only validates the input data once at the start rather than at every call to Connect. However, the end
// result is the same as doing so.
func (b *Breadboard) ConnectGroup(head, tail ID) error {
	if !b.exists(head) || !b.exists(tail) {
		return ErrInvalidID
	}
	if len(b.groups[head].pins) != len(b.groups[tail].pins) {
		return ErrNonUniformGroups
	}
	for i := range len(b.groups[head].pins) {
		b.connect(Wire{
			Head: Pin{
				ID:    head,
				Index: i,
			},
			Tail: Pin{
				ID:    tail,
				Index: i,
			},
		})
	}
	return nil
}

// Set registers the provided value to the provided [hdl.Pin]. If the provided value is different from the current value
// of the pin then the registered callback for the pin is executed.
func (b *Breadboard) Set(pin Pin, value byte) {
	if err := b.validate(pin); err != nil {
		panic(err)
	}
	if b.Get(pin) == value {
		return
	}
	b.set(pin, value)
}

// SetGroup works closely to [hdl.Breadboard.Set] but instead of setting a single pin it sets a whole group of groups in a
// single method call. For each of the new values set on a [hdl.Pin] within the provided group the registered callback
// is invoked. This behaviour is subject to change though and will likely be revamped such that
// [hdl.Breadboard.SetGroup] only invokes the callback once.
func (b *Breadboard) SetGroup(id ID, values []byte) error {
	if !b.exists(id) {
		return ErrInvalidID
	}
	if len(b.groups[id].pins) != len(values) {
		return ErrNonUniformGroups
	}
	for i := range values {
		pin := Pin{
			ID:    id,
			Index: i,
		}
		if b.Get(pin) == values[i] {
			continue
		}
		b.set(pin, values[i])
	}
	return nil
}

func (b *Breadboard) set(pin Pin, value byte) {
	b.changeset.enqueue(pin)
	b.groups[pin.ID].pins[pin.Index].set(value)
}

func (b *Breadboard) Get(pin Pin) byte {
	if err := b.validate(pin); err != nil {
		panic(err)
	}
	return b.groups[pin.ID].pins[pin.Index].value
}

func (b *Breadboard) GetGroup(id ID) ([]byte, error) {
	if !b.exists(id) {
		return nil, ErrInvalidID
	}
	pins := make([]byte, len(b.groups[id].pins))
	for i := range b.groups[id].pins {
		pins[i] = b.groups[id].pins[i].value
	}
	return pins, nil
}

func (b *Breadboard) validate(pin Pin) error {
	if ok := b.exists(pin.ID); !ok {
		return ErrInvalidID
	}
	if pin.Index > len(b.groups[pin.ID].pins) {
		return ErrInvalidIndex
	}
	return nil
}

func (b *Breadboard) exists(id ID) bool {
	if id < 0 || id > len(b.groups) {
		return false
	}
	return true
}

func Tick(b *Breadboard) {
	b.Set(Pin{ID: b.CLK, Index: 0}, 1)
	for b.changeset.more() {
		pin := b.changeset.dequeue()
		pins, _ := b.GetGroup(pin.ID)
		g := b.groups[pin.ID]
		if g.callback != nil {
			g.callback(pin.ID, pins)
		}
		children, ok := b.wires[pin]
		if !ok {
			continue
		}
		for _, child := range children {
			b.set(child, pins[pin.Index])
		}
	}
	b.Set(Pin{ID: b.CLK, Index: 0}, 0)
}
