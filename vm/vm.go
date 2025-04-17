package vm

import (
	"errors"
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
	vm := VM{lx: lx, file: file}
	return vm.Translate()
}

type VM struct {
	file     string
	context  string
	lx       internal.Lexer
	sequence int
	statics  [240]int
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

func (vm *VM) Translate() ([]string, error) {
	commands := make([]Command, 0)
	seq := int32(0)
	for vm.lx.More() {
		cmd, err := vm.parseCommand()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
		seq += 1
	}
	assembly := make([]string, 0)
	for _, command := range commands {
		a, err := command.Evaluate()
		if err != nil {
			return nil, err
		}
		assembly = append(assembly, a...)
	}
	return assembly, nil
}

func (vm *VM) parseCommand() (Command, error) {
	token, err := vm.lx.Next()
	if err != nil {
		return nil, err
	}
	switch token.Variant {
	case internal.Push:
		return vm.parsePushCommand()
	case internal.Pop:
		return vm.parsePopCommand()
	case internal.Add:
		return CommandFunc(AddCommand), nil
	case internal.Sub:
		return CommandFunc(SubCommand), nil
	case internal.Neg:
		return CommandFunc(NegCommand), nil
	case internal.And:
		return CommandFunc(AndCommand), nil
	case internal.Or:
		return CommandFunc(OrCommand), nil
	case internal.Not:
		return CommandFunc(NotCommand), nil
	case internal.Eq:
		return EqCommand(vm.seq()), nil
	case internal.Lt:
		return LtCommand(vm.seq()), nil
	case internal.Gt:
		return GtCommand(vm.seq()), nil
	case internal.Function:
		name, err := vm.expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		vars, err := vm.parseInteger()
		if err != nil {
			return nil, err
		}
		vm.context = name.Literal
		return FunctionCommand(fmt.Sprintf("%s.%s", vm.file, vm.context), vars), nil
	case internal.Return:
		return CommandFunc(ReturnCommand), nil
	case internal.Call:
		callee, err := vm.expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		args, err := vm.parseInteger()
		if err != nil {
			return nil, err
		}
		caller := fmt.Sprintf("%s.%s", vm.file, vm.context)
		return CallCommand(caller, callee.Literal, args, vm.seq()), nil
	case internal.Label:
		name, err := vm.expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		return CommandFunc(func() ([]string, error) {
			return []string{fmt.Sprintf("(%s)", name.Literal)}, nil
		}), nil
	case internal.Goto:
		name, err := vm.expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		return CommandFunc(func() ([]string, error) {
			return []string{
				fmt.Sprintf("@%s", name.Literal),
				"0;JMP",
			}, nil
		}), nil
	case internal.IfGoto:
		label, err := vm.expect(internal.Identifier)
		if err != nil {
			return nil, err
		}
		return CommandFunc(func() ([]string, error) {
			return write(
				CommandFunc(Pop),
				Constant(
					"D=M",
					fmt.Sprintf("@%s", label.Literal),
					"D;JGT",
				),
			)
		}), nil
	default:
		panic(fmt.Errorf("variant %v is not yet supported", token.Variant))
	}
}

func (vm *VM) seq() int {
	vm.sequence++
	return vm.sequence
}

func (vm *VM) parsePushCommand() (Command, error) {
	src, err := vm.lx.Next()
	if err != nil {
		return nil, err
	}
	index, err := vm.parseInteger()
	if err != nil {
		return nil, err
	}
	switch src.Variant {
	case internal.Constant:
		return PushCommand(ReadConstant(index)), nil
	case internal.Tmp:
		return PushCommand(ReadTemp(index)), nil
	case internal.Pointer:
		return PushCommand(ReadPointer(index)), nil
	case internal.Static:
		return PushCommand(ReadMemory(vm.static(index))), nil
	case internal.Local:
		return PushCommand(ReadSegment("LCL", index)), nil
	case internal.Arg:
		return PushCommand(ReadSegment("ARG", index)), nil
	case internal.This:
		return PushCommand(ReadSegment("THIS", index)), nil
	case internal.That:
		return PushCommand(ReadSegment("THAT", index)), nil
	default:
		return nil, fmt.Errorf("unexpected variant: %v", src.Variant)
	}
}

func (vm *VM) parsePopCommand() (Command, error) {
	target, err := vm.lx.Next()
	if err != nil {
		return nil, err
	}
	index, err := vm.parseInteger()
	if err != nil {
		return nil, err
	}
	switch target.Variant {
	case internal.Tmp:
		return PopCommand(TempTarget(index)), nil
	case internal.Pointer:
		return PopCommand(PointerTarget(index)), nil
	case internal.Static:
		return PopCommand(MemoryTarget(vm.static(index))), nil
	case internal.Local:
		return PopCommand(SegmentTarget("LCL", index)), nil
	case internal.Arg:
		return PopCommand(SegmentTarget("ARG", index)), nil
	case internal.This:
		return PopCommand(SegmentTarget("THIS", index)), nil
	case internal.That:
		return PopCommand(SegmentTarget("THAT", index)), nil
	default:
		return nil, fmt.Errorf("unexpected variant: %v", target.Variant)
	}
}

func (vm *VM) parseInteger() (int, error) {
	token, err := vm.expect(internal.Integer)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(token.Literal)
}

func (vm *VM) expect(v internal.Variant) (internal.Token, error) {
	token, err := vm.lx.Next()
	if err != nil {
		return internal.NullToken, err
	}
	if token.Variant != v {
		return internal.NullToken, fmt.Errorf("expected variant %v but found %v", v, token.Variant)
	}
	return token, nil
}
