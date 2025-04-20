package vm

import (
	"fmt"
	"github.com/crookdc/nand2tetris/vm/internal"
	"io"
	"strconv"
)

func Translate(file string, r io.Reader) ([]string, error) {
	lx, err := internal.NewLexer(r)
	if err != nil {
		return nil, err
	}
	vm := VM{
		file: file,
		lx:   lx,
	}
	cmds := make(Command, 0)
	for cmd, err := vm.Next(); err != io.EOF; cmd, err = vm.Next() {
		cmds = append(cmds, cmd...)
	}
	tree, err := cmds.Compile()
	if err != nil {
		return nil, err
	}
	asm := make([]string, 0)
	for _, ins := range tree {
		asm = append(asm, ins.Get())
	}
	return asm, nil
}

type Command []AssemblyGenerator

func (c Command) Compile() ([]AssemblyInstruction, error) {
	asm := make([]AssemblyInstruction, 0)
	for _, a := range c {
		cmp, err := a.Get()
		if err != nil {
			return nil, err
		}
		asm = append(asm, cmp...)
	}
	return asm, nil
}

type VM struct {
	file     string
	context  string
	lx       internal.Lexer
	sequence int
	statics  [240]int
}

func (vm *VM) Next() (Command, error) {
	token, err := vm.lx.Next()
	if err != nil {
		return nil, err
	}
	switch token.Variant {
	case internal.Push:
		t, err := vm.lx.Next()
		if err != nil {
			return nil, err
		}
		i, err := vm.nextInteger()
		if err != nil {
			return nil, err
		}
		return vm.Push(t, i), nil
	case internal.Pop:
		t, err := vm.lx.Next()
		if err != nil {
			return nil, err
		}
		i, err := vm.nextInteger()
		if err != nil {
			return nil, err
		}
		return vm.Pop(t, i), nil
	case internal.Add:
		return vm.Add(), nil
	case internal.Sub:
		return vm.Sub(), nil
	case internal.Neg:
		return vm.Neg(), nil
	case internal.And:
		return vm.And(), nil
	case internal.Or:
		return vm.Or(), nil
	case internal.Not:
		return vm.Not(), nil
	case internal.Eq:
		return vm.Eq(), nil
	case internal.Lt:
		return vm.Lt(), nil
	case internal.Gt:
		return vm.Gt(), nil
	case internal.Goto:
		l, err := vm.lx.Expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		return vm.Goto(l.Literal), nil
	case internal.IfGoto:
		l, err := vm.lx.Expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		return vm.IfGoto(l.Literal), nil
	case internal.Return:
		return vm.Return(), nil
	case internal.Function:
		fn, err := vm.lx.Expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		nArgs, err := vm.nextInteger()
		if err != nil {
			return nil, err
		}
		vm.context = fn.Literal
		return vm.Function(fn.Literal, nArgs), nil
	case internal.Call:
		fn, err := vm.lx.Expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		nArgs, err := vm.nextInteger()
		if err != nil {
			return nil, err
		}
		return vm.Call(fn.Literal, nArgs), nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Literal)
	}
}

func (vm *VM) nextInteger() (int, error) {
	i, err := vm.lx.Expect(internal.Integer)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(i.Literal)
}

func (vm *VM) static(index int) int {
	if vm.statics[index] > 15 {
		// The references static has already been allocated in memory
		return vm.statics[index]
	}
	// Find the next available memory address and allocate the index to it
	current := 15
	for _, n := range vm.statics {
		if n > current {
			current = n
		}
	}
	vm.statics[index] = current + 1
	return vm.statics[index]
}

func (vm *VM) Add() Command {
	return vm.binary(Add[D]{
		X: D{},
		Y: M{},
	})
}

func (vm *VM) Sub() Command {
	return vm.binary(Sub[M]{
		X: M{},
		Y: D{},
	})
}

func (vm *VM) And() Command {
	return vm.binary(And[D, M]{
		X: D{},
		Y: M{},
	})
}

func (vm *VM) Or() Command {
	return vm.binary(Or[D, M]{
		X: D{},
		Y: M{},
	})
}

func (vm *VM) binary(operation AssemblyInstruction) Command {
	return []AssemblyGenerator{
		Pop{},
		InlineGenerator(
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
		),
		Pop{},
		InlineGenerator(
			Assign{
				Source:  operation,
				Targets: Target(D{}),
			},
		),
		Push{},
	}
}

func (vm *VM) Neg() Command {
	return vm.unary(Negate[M]{X: M{}})
}

func (vm *VM) Not() Command {
	return vm.unary(Not[M]{X: M{}})
}

func (vm *VM) unary(operation AssemblyInstruction) Command {
	return []AssemblyGenerator{
		Pop{},
		InlineGenerator(
			Assign{
				Source:  operation,
				Targets: Target(D{}),
			},
		),
		Push{},
	}
}

func (vm *VM) Push(t internal.Token, index int) Command {
	// The preload is responsible for getting the data that should be pushed to the stack to the D register
	var preload AssemblyGenerator
	switch t.Variant {
	case internal.Constant:
		preload = LoadInto{
			Value:   strconv.Itoa(index),
			Targets: Target(D{}),
		}
	case internal.Local:
		preload = LoadIndexedMemory{
			IndexedMemory{
				Base:  "LCL",
				Index: index,
			},
		}
	case internal.Argument:
		preload = LoadIndexedMemory{
			IndexedMemory{
				Base:  "ARG",
				Index: index,
			},
		}
	case internal.This:
		preload = LoadIndexedMemory{
			IndexedMemory{
				Base:  "THIS",
				Index: index,
			},
		}
	case internal.That:
		preload = LoadIndexedMemory{
			IndexedMemory{
				Base:  "THAT",
				Index: index,
			},
		}
	case internal.Temp:
		preload = LoadTempInto{
			Index:   index,
			Targets: Target(D{}),
		}
	case internal.Static:
		preload = LoadInto{
			Value:   strconv.Itoa(vm.static(index)),
			Targets: Target(D{}),
		}
	default:
		panic(fmt.Errorf("push type not supported: %s", t.Literal))
	}
	return []AssemblyGenerator{
		preload,
		Push{},
	}
}

func (vm *VM) Pop(t internal.Token, index int) Command {
	var preload AssemblyGenerator
	switch t.Variant {
	case internal.Local:
		preload = InlineGenerator(Load{Value: "LCL"})
	case internal.Argument:
		preload = InlineGenerator(Load{Value: "ARG"})
	case internal.This:
		preload = InlineGenerator(Load{Value: "THIS"})
	case internal.That:
		preload = InlineGenerator(Load{Value: "THAT"})
	case internal.Temp:
		preload = InlineGenerator(Load{Value: "5"})
	case internal.Static:
		preload = InlineGenerator(Load{Value: strconv.Itoa(vm.static(index))})
	default:
		panic(fmt.Errorf("pop target not supported: %s", t.Literal))
	}
	return []AssemblyGenerator{
		Pop{},
		InlineGenerator(
			Assign{Source: M{}, Targets: Target(D{})},
		),
		preload,
		InlineGenerator(
			Assign{
				Source: Add[D]{
					X: D{},
					Y: M{},
				},
				Targets: Target(D{}),
			},
			Load{Value: strconv.Itoa(index)},
			Assign{
				Source: Add[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(D{}),
			},
			Load{Value: StackPointer},
			Assign{
				Source:  M{},
				Targets: Target(A{}),
			},
			Assign{
				Source:  M{},
				Targets: Target(A{}),
			},
			Assign{
				Source: Sub[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(A{}),
			},
			Assign{
				Source: Sub[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(M{}),
			},
		),
	}
}

func (vm *VM) Goto(label string) Command {
	return []AssemblyGenerator{
		InlineGenerator(
			Load{Value: label},
			JMP{Instruction: Zero{}},
		),
	}
}

func (vm *VM) IfGoto(label string) Command {
	return []AssemblyGenerator{
		Pop{},
		InlineGenerator(
			Assign{Source: M{}, Targets: Target(D{})},
			Load{Value: label},
			JGT{Instruction: D{}},
		),
	}
}

func (vm *VM) Eq() Command {
	return vm.comparison(JEQ{Instruction: D{}})
}

func (vm *VM) Lt() Command {
	return vm.comparison(JLT{Instruction: D{}})
}

func (vm *VM) Gt() Command {
	return vm.comparison(JGT{Instruction: D{}})
}

func (vm *VM) comparison(jmp AssemblyInstruction) Command {
	seq := vm.seq()
	return []AssemblyGenerator{
		Pop{},
		InlineGenerator(
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
		),
		Pop{},
		InlineGenerator(
			Assign{
				Source: Sub[M]{
					X: M{},
					Y: D{},
				},
				Targets: Target(D{}),
			},
		),
		InlineGenerator(
			Load{Value: fmt.Sprintf("T.%d", seq)},
			jmp,
			Assign{
				Source:  Zero{},
				Targets: Target(D{}),
			},
			Load{Value: fmt.Sprintf("END.%d", seq)},
			JMP{Instruction: Zero{}},
			Label{Value: fmt.Sprintf("T.%d", seq)},
			Assign{
				Source:  One{Negative: true},
				Targets: Target(D{}),
			},
			Label{Value: fmt.Sprintf("END.%d", seq)},
		),
		Push{},
	}
}

func (vm *VM) seq() int {
	vm.sequence++
	return vm.sequence
}

func (vm *VM) Function(name string, nArgs int) Command {
	cmd := []AssemblyGenerator{
		InlineGenerator(
			Label{Value: fmt.Sprintf("%s.%s", vm.file, name)},
		),
	}
	for range nArgs {
		cmd = append(
			cmd,
			InlineGenerator(
				Load{Value: "0"},
				Assign{
					Source:  A{},
					Targets: Target(D{}),
				},
			),
			Push{},
		)
	}
	return cmd
}

func (vm *VM) Return() Command {
	rewind := func(segment string, offset int) AssemblyGenerator {
		return InlineGenerator(
			Load{Value: "R13"},
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
			Load{Value: strconv.Itoa(offset)},
			Assign{
				Source: Sub[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(A{}, D{}),
			},
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
		)
	}
	return []AssemblyGenerator{
		InlineGenerator(
			// Stores LCL (frame) in R13
			Load{Value: "LCL"},
			Assign{Source: M{}, Targets: Target(D{})},
			Load{Value: "R13"},
			Assign{Source: D{}, Targets: Target(M{})},
			// Stores return address in R14
			Load{Value: "5"},
			Assign{
				Source: Sub[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(A{}, D{}),
			},
			Load{Value: "R14"},
			Assign{
				Source:  D{},
				Targets: Target(M{}),
			},
		),
		// Pop stack into ARG
		Pop{},
		InlineGenerator(
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
			Load{Value: "ARG"},
			Assign{
				Source:  M{},
				Targets: Target(A{}),
			},
			Assign{
				Source:  D{},
				Targets: Target(M{}),
			},
			Load{Value: "ARG"},
			Assign{
				Source:  D{},
				Targets: Target(M{}),
			},
		),
		InlineGenerator(
			// Reposition stack pointer
			Load{Value: StackPointer},
			Assign{
				Source: Add[D]{
					X: D{},
					Y: One{},
				},
				Targets: Target(M{}),
			},
		),
		rewind("THAT", 1),
		rewind("THIS", 2),
		rewind("ARG", 3),
		rewind("LCL", 4),
		InlineGenerator(
			Load{Value: "R14"},
			Assign{
				Source:  M{},
				Targets: Target(A{}),
			},
			JMP{Instruction: Zero{}},
		),
	}
}

func (vm *VM) Call(fn string, nArgs int) Command {
	retAddr := fmt.Sprintf("%s.%s$ret%d", vm.file, vm.context, vm.seq())
	return []AssemblyGenerator{
		LoadInto{
			Value:   retAddr,
			Targets: Target(D{}),
		},
		Push{},
		LoadMemoryInto{
			Value:   "LCL",
			Targets: Target(D{}),
		},
		Push{},
		LoadMemoryInto{
			Value:   "ARG",
			Targets: Target(D{}),
		},
		Push{},
		LoadMemoryInto{
			Value:   "THIS",
			Targets: Target(D{}),
		},
		Push{},
		LoadMemoryInto{
			Value:   "THAT",
			Targets: Target(D{}),
		},
		Push{},
		InlineGenerator(
			Load{Value: StackPointer},
			Assign{
				Source:  M{},
				Targets: Target(D{}),
			},
			Load{Value: strconv.Itoa(5 + nArgs)},
			Assign{
				Source: Sub[D]{
					X: D{},
					Y: A{},
				},
				Targets: Target(D{}),
			},
			Load{Value: "ARG"},
			Assign{
				Source:  D{},
				Targets: Target(M{}),
			},
		),
		InlineGenerator(
			Load{Value: fn},
			JMP{Instruction: Zero{}},
			Label{Value: retAddr},
		),
	}
}
