package vm

import (
	"fmt"
	"strconv"
)

var (
	StackPointer = "SP"
)

type Targets []AssemblyInstruction

func (t Targets) Get() string {
	if len(t) < 1 || len(t) > 3 {
		panic(fmt.Errorf("invalid target amount: %d", len(t)))
	}
	dst := ""
	for _, tgt := range t {
		switch tgt.(type) {
		case A, D, M:
			dst += tgt.Get()
		default:
			panic(fmt.Errorf("invalid target: %s", tgt.Get()))
		}
	}
	return dst
}

func Target(targets ...AssemblyInstruction) Targets {
	for _, t := range targets {
		switch t.(type) {
		case A, D, M:
			continue
		default:
			panic(fmt.Errorf("invalid target: %s", t.Get()))
		}
	}
	return targets
}

type AssemblyInstruction interface {
	Get() string
}

type A struct{}

func (a A) Get() string {
	return "A"
}

type D struct{}

func (d D) Get() string {
	return "D"
}

type M struct{}

func (m M) Get() string {
	return "M"
}

type Assignable interface {
	AssemblyInstruction
	A | D | M
}

type One struct {
	Negative bool
}

func (o One) Get() string {
	if o.Negative {
		return "-1"
	}
	return "1"
}

type Zero struct{}

func (z Zero) Get() string {
	return "0"
}

type Load struct {
	Value string
}

func (l Load) Get() string {
	return fmt.Sprintf("@%s", l.Value)
}

type Label struct {
	Value string
}

func (l Label) Get() string {
	return fmt.Sprintf("(%s)", l.Value)
}

type JMP struct {
	Instruction AssemblyInstruction
}

func (j JMP) Get() string {
	return fmt.Sprintf("%s;JMP", j.Instruction.Get())
}

type JEQ struct {
	Instruction AssemblyInstruction
}

func (j JEQ) Get() string {
	return fmt.Sprintf("%s;JEQ", j.Instruction.Get())
}

type JLT struct {
	Instruction AssemblyInstruction
}

func (j JLT) Get() string {
	return fmt.Sprintf("%s;JLT", j.Instruction.Get())
}

type JGT struct {
	Instruction AssemblyInstruction
}

func (j JGT) Get() string {
	return fmt.Sprintf("%s;JGT", j.Instruction.Get())
}

type Assign struct {
	Source  AssemblyInstruction
	Targets Targets
}

func (a Assign) Get() string {
	return fmt.Sprintf("%s=%s", a.Targets.Get(), a.Source.Get())
}

type Sub[X Assignable] struct {
	X X
	Y AssemblyInstruction
}

func (s Sub[X]) Get() string {
	return fmt.Sprintf("%s-%s", s.X.Get(), s.Y.Get())
}

type Add[X Assignable] struct {
	X X
	Y AssemblyInstruction
}

func (a Add[X]) Get() string {
	return fmt.Sprintf("%s+%s", a.X.Get(), a.Y.Get())
}

type And[X, Y Assignable] struct {
	X X
	Y Y
}

func (a And[X, Y]) Get() string {
	return fmt.Sprintf("%s&%s", a.X.Get(), a.Y.Get())
}

type Or[X, Y Assignable] struct {
	X X
	Y Y
}

func (a Or[X, Y]) Get() string {
	return fmt.Sprintf("%s|%s", a.X.Get(), a.Y.Get())
}

type Not[X Assignable] struct {
	X X
}

func (n Not[X]) Get() string {
	return fmt.Sprintf("!%s", n.X.Get())
}

type Negate[X Assignable] struct {
	X X
}

func (n Negate[X]) Get() string {
	return fmt.Sprintf("-%s", n.X.Get())
}

type AssemblyGenerator interface {
	Get() ([]AssemblyInstruction, error)
}

func InlineGenerator(instruction ...AssemblyInstruction) AssemblyGeneratorFunc {
	return func() ([]AssemblyInstruction, error) {
		return instruction, nil
	}
}

type AssemblyGeneratorFunc func() ([]AssemblyInstruction, error)

func (a AssemblyGeneratorFunc) Get() ([]AssemblyInstruction, error) {
	return a()
}

type Push struct{}

func (p Push) Get() ([]AssemblyInstruction, error) {
	return []AssemblyInstruction{
		Load{Value: StackPointer},
		Assign{
			Source:  M{},
			Targets: Target(A{}),
		},
		Assign{
			Source:  D{},
			Targets: Target(M{}),
		},
		Load{Value: StackPointer},
		Assign{
			Source: Add[M]{
				X: M{},
				Y: One{},
			},
			Targets: Target(M{}),
		},
	}, nil
}

type Pop struct{}

func (p Pop) Get() ([]AssemblyInstruction, error) {
	return []AssemblyInstruction{
		Load{Value: StackPointer},
		Assign{
			Source: Sub[M]{
				X: M{},
				Y: One{},
			},
			Targets: Target(A{}, M{}),
		},
	}, nil
}

type LoadInto struct {
	Value string
	Targets
}

func (l LoadInto) Get() ([]AssemblyInstruction, error) {
	return []AssemblyInstruction{
		Load{Value: l.Value},
		Assign{
			Source:  A{},
			Targets: l.Targets,
		},
	}, nil
}

type LoadMemoryInto struct {
	Value string
	Targets
}

func (l LoadMemoryInto) Get() ([]AssemblyInstruction, error) {
	return []AssemblyInstruction{
		Load{Value: l.Value},
		Assign{
			Source:  M{},
			Targets: l.Targets,
		},
	}, nil
}

type LoadTempInto struct {
	Index int
	Targets
}

func (l LoadTempInto) Get() ([]AssemblyInstruction, error) {
	if l.Index < 0 || l.Index > 7 {
		return nil, fmt.Errorf("invalid temp index %d", l.Index)
	}
	index := l.Index + 5
	return []AssemblyInstruction{
		Load{Value: strconv.Itoa(index)},
		Assign{
			Source:  M{},
			Targets: l.Targets,
		},
	}, nil
}

type IndexedMemory struct {
	Base  string
	Index int
}

type LoadIndexedMemory struct {
	IndexedMemory
}

func (i LoadIndexedMemory) Get() ([]AssemblyInstruction, error) {
	return []AssemblyInstruction{
		Load{Value: i.Base},
		Assign{
			Source:  M{},
			Targets: Target(D{}),
		},
		Load{Value: strconv.Itoa(i.Index)},
		Assign{
			Source: Add[D]{
				X: D{},
				Y: A{},
			},
			Targets: Target(A{}),
		},
		Assign{
			Source:  M{},
			Targets: Target(D{}),
		},
	}, nil
}
