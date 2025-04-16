package vm

import (
	"errors"
	"fmt"
	"github.com/crookdc/nand2tetris/lexer"
	"io"
	"strconv"
)

var segments = map[variant]string{
	local: "LCL",
	arg:   "ARG",
	this:  "THIS",
	that:  "THAT",
}

func Translate(file string, r io.Reader) ([]string, error) {
	lx, err := newLexer(r)
	if err != nil {
		return nil, err
	}
	vm := VM{lx: lx, file: file}
	return vm.Translate()
}

type VM struct {
	file     string
	context  string
	lx       *lexer.Lexer[variant]
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
	case push:
		return vm.parsePushCommand()
	case pop:
		return vm.parsePopCommand()
	case add:
		return CommandFunc(AddCommand), nil
	case sub:
		return CommandFunc(SubCommand), nil
	case neg:
		return CommandFunc(NegCommand), nil
	case and:
		return CommandFunc(AndCommand), nil
	case or:
		return CommandFunc(OrCommand), nil
	case not:
		return CommandFunc(NotCommand), nil
	case eq:
		return EqCommand(vm.seq()), nil
	case lt:
		return LtCommand(vm.seq()), nil
	case gt:
		return GtCommand(vm.seq()), nil
	case fn:
		name, err := vm.expect(identifier)
		if err != nil {
			return nil, err
		}
		vars, err := vm.parseInteger()
		if err != nil {
			return nil, err
		}
		vm.context = name.Literal
		return FunctionCommand(fmt.Sprintf("%s.%s", vm.file, vm.context), vars), nil
	case ret:
		return CommandFunc(ReturnCommand), nil
	case call:
		callee, err := vm.expect(identifier)
		if err != nil {
			return nil, err
		}
		args, err := vm.parseInteger()
		if err != nil {
			return nil, err
		}
		caller := fmt.Sprintf("%s.%s", vm.file, vm.context)
		return CallCommand(caller, callee.Literal, args, vm.seq()), nil
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
	case constant:
		return PushCommand(ReadConstant(index)), nil
	case tmp:
		return PushCommand(ReadTemp(index)), nil
	case pointer:
		return PushCommand(ReadPointer(index)), nil
	case static:
		return PushCommand(ReadMemory(vm.static(index))), nil
	default:
		segment, ok := segments[src.Variant]
		if !ok {
			return nil, fmt.Errorf("unexpected push source %v", src.Literal)
		}
		return PushCommand(ReadSegment(segment, index)), nil
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
	case tmp:
		return PopCommand(TempTarget(index)), nil
	case pointer:
		return PopCommand(PointerTarget(index)), nil
	case static:
		return PopCommand(MemoryTarget(vm.static(index))), nil
	default:
		segment, ok := segments[target.Variant]
		if !ok {
			return nil, fmt.Errorf("unexpected pop target %v", target.Literal)
		}
		return PopCommand(SegmentTarget(segment, index)), nil
	}
}

func (vm *VM) parseInteger() (int, error) {
	token, err := vm.expect(integer)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(token.Literal)
}

func (vm *VM) expect(v variant) (lexer.Token[variant], error) {
	token, err := vm.lx.Next()
	if err != nil {
		return lexer.Token[variant]{}, err
	}
	if token.Variant != v {
		return lexer.Token[variant]{}, fmt.Errorf("expected variant %v but found %v", v, token.Variant)
	}
	return token, nil
}
