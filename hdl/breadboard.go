package hdl

import "errors"

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidIndex     = errors.New("index out of range")
	ErrNonUniformGroups = errors.New("groups are not uniform")
)

var Noop = func(ID, []byte) {}

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
	return &Breadboard{
		pins:      make([][]signal, 0),
		wires:     make(map[Pin][]Pin),
		callbacks: make(map[ID]Callback),
	}
}

type Breadboard struct {
	pins      [][]signal
	wires     map[Pin][]Pin
	callbacks map[ID]Callback
}

// SizeOf returns the length of the group registered under the provided ID. An error is returned if the ID is not
// registered on the Breadboard or is otherwise invalid.
func (b *Breadboard) SizeOf(id ID) (int, error) {
	if !b.exists(id) {
		return 0, ErrInvalidID
	}
	return len(b.pins[id]), nil
}

// Allocate allocates a new pin group of the provided size together with a callback to be called when a new value is set
// on any of the pins within the group. The returned ID is the ID to be used when retrieving or setting values within
// the group. See [hdl.Breadboard.Set], [hdl.Breadboard.Get], [hdl.Breadboard.Get] and [hdl.Breadboard.GetGroup].
func (b *Breadboard) Allocate(count int, cb Callback) ID {
	id := len(b.pins)
	b.pins = append(b.pins, make([]signal, count))
	b.callbacks[id] = cb
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
	b.set(wire.Head, b.Get(wire.Head))
}

// ConnectGroup is a convenience method that connects all pins of the groups identified by the supplied ID's with head
// as the driver. It is negligibly quicker than calling Connect several times while iterating over a known number of
// pins since it only validates the input data once at the start rather than at every call to Connect. However, the end
// result is the same as doing so.
func (b *Breadboard) ConnectGroup(head, tail ID) error {
	if !b.exists(head) {
		return ErrInvalidID
	}
	if !b.exists(tail) {
		return ErrInvalidID
	}
	if len(b.pins[head]) != len(b.pins[tail]) {
		return ErrNonUniformGroups
	}
	for i := range len(b.pins[head]) {
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

// SetGroup works closely to [hdl.Breadboard.Set] but instead of setting a single pin it sets a whole group of pins in a
// single method call. For each of the new values set on a [hdl.Pin] within the provided group the registered callback
// is invoked. This behaviour is subject to change though and will likely be revamped such that
// [hdl.Breadboard.SetGroup] only invokes the callback once.
func (b *Breadboard) SetGroup(id ID, values []byte) error {
	if !b.exists(id) {
		return ErrInvalidID
	}
	if len(b.pins[id]) != len(values) {
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
		// If optimization is needed when setting groups this is the first place to revisit. Just calling set for each
		// pin is rather clean, but it does run the callback for each pin being set rather than just once at the end
		// after all pins have been set.
		b.set(pin, values[i])
	}
	return nil
}

func (b *Breadboard) set(pin Pin, value byte) {
	b.pins[pin.ID][pin.Index].set(value)
	pins, _ := b.GetGroup(pin.ID)
	b.callbacks[pin.ID](pin.ID, pins)
	children, ok := b.wires[pin]
	if !ok {
		return
	}
	for _, child := range children {
		b.set(child, value)
	}
}

func (b *Breadboard) Get(pin Pin) byte {
	if err := b.validate(pin); err != nil {
		panic(err)
	}
	return b.pins[pin.ID][pin.Index].value
}

func (b *Breadboard) GetGroup(id ID) ([]byte, error) {
	if !b.exists(id) {
		return nil, ErrInvalidID
	}
	pins := make([]byte, len(b.pins[id]))
	for i := range b.pins[id] {
		pins[i] = b.pins[id][i].value
	}
	return pins, nil
}

func (b *Breadboard) validate(pin Pin) error {
	if ok := b.exists(pin.ID); !ok {
		return ErrInvalidID
	}
	if pin.Index > len(b.pins[pin.ID]) {
		return ErrInvalidIndex
	}
	return nil
}

func (b *Breadboard) exists(id ID) bool {
	if id < 0 || id > len(b.pins) {
		return false
	}
	return true
}
