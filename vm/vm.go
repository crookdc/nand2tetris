package vm

import (
	"errors"
	"fmt"
	"github.com/crookdc/nand2tetris/lexer"
	"io"
	"strconv"
)

var ()

type Command interface {
	Evaluate() ([]string, error)
}

type CommandFunc func() ([]string, error)

func (c CommandFunc) Evaluate() ([]string, error) {
	return c()
}

func ReadConstant(value int) CommandFunc {
	return func() ([]string, error) {
		if value < 0 || value > 32_767 {
			return nil, fmt.Errorf("invalid constant value %d", value)
		}
		return []string{
			fmt.Sprintf("@%d", value),
			"D=A",
		}, nil
	}
}

func ReadSegment(sgm string, index int) CommandFunc {
	return func() ([]string, error) {
		if index < 0 {
			return nil, fmt.Errorf("invalid segment index %d", index)
		}
		return []string{
			fmt.Sprintf("@%s", sgm),
			"D=M",
			fmt.Sprintf("@%d", index),
			"A=D+A",
			"D=M",
		}, nil
	}
}

func PushCommand(src Command) CommandFunc {
	return func() ([]string, error) {
		return write(
			src,
			CommandFunc(PushStack),
		)
	}
}

type PopCommand struct {
	target Command
}

func IncrementStack() ([]string, error) {
	return []string{
		"@SP",
		"M=M+1",
	}, nil
}

func DecrementStack() ([]string, error) {
	return []string{
		"@SP",
		"M=M-1",
	}, nil
}

func PopStack() ([]string, error) {
	return []string{
		"@SP",   // Load stack pointer segment
		"M=M-1", // Decrement address of the stack pointer
		"A=M",   // Set the current memory address to stack pointer value
		"D=M",   // Grab the value at the address and place it in D
	}, nil
}

func WriteVirtual(src, target string) CommandFunc {
	return func() ([]string, error) {
		return []string{
			"@" + target,
			"M=" + src,
		}, nil
	}
}

func PushStack() ([]string, error) {
	return []string{
		"@SP",   // Load stack pointer segment
		"A=M",   // Set the current memory address to stack pointer value
		"M=D",   // Grab the value at the address and place it in D
		"@SP",   //
		"M=M+1", // Increment stack pointer
	}, nil
}

func Constant(cmd ...string) CommandFunc {
	return func() ([]string, error) {
		return cmd, nil
	}
}

func AddCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		WriteVirtual("D", "R13"),
		CommandFunc(PopStack),
		WriteVirtual("D", "R14"),
		Constant(
			"@R13",
			"D=M",
			"@R14",
			"D=D+M",
		),
	)
}

func write(commands ...Command) ([]string, error) {
	asm := make([]string, 0)
	for _, c := range commands {
		n, err := c.Evaluate()
		if err != nil {
			return nil, err
		}
		asm = append(asm, n...)
	}
	return asm, nil
}

func Evaluate(r io.Reader) ([]string, error) {
	l, err := newLexer(r)
	if err != nil {
		return nil, err
	}
	commands := make([]Command, 0)
	for l.More() {
		cmd, err := parse(l)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
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

func parse(l *lexer.Lexer[variant]) (Command, error) {
	next, err := l.Next()
	if err != nil {
		return nil, err
	}
	switch next.Variant {
	case push:
		segment, err := l.Next()
		if err != nil {
			return nil, err
		}
		index, err := l.Next()
		if err != nil {
			return nil, err
		}
		if index.Variant != integer {
			return nil, fmt.Errorf("unexpected token variant %v", index.Variant)
		}
		parsed, err := strconv.Atoi(index.Literal)
		if err != nil {
			return nil, fmt.Errorf("cannot parse integer: %w", err)
		}
		switch segment.Variant {
		case constant:
			return PushCommand(ReadConstant(parsed)), nil
		case local:
			return PushCommand(ReadSegment("LCL", parsed)), nil
		case arg:
			return PushCommand(ReadSegment("ARG", parsed)), nil
		case this:
			return PushCommand(ReadSegment("THIS", parsed)), nil
		case that:
			return PushCommand(ReadSegment("THAT", parsed)), nil
		default:
			panic("not implemented")
		}
	default:
		panic(fmt.Errorf("variant %v is not yet supported", next.Variant))
	}
}

func expect(l *lexer.Lexer[variant], expected variant) (lexer.Token[variant], error) {
	token, err := l.Next()
	if err != nil {
		return lexer.Token[variant]{}, err
	}
	if token.Variant != expected {
		return lexer.Token[variant]{}, fmt.Errorf("expected token variant %v but got %v", expected, token.Variant)
	}
	return token, nil
}
