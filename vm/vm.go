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

func SegmentTarget(segment string, index int) CommandFunc {
	return func() ([]string, error) {
		if index < 0 {
			return nil, fmt.Errorf("invalid segment index %d", index)
		}
		return []string{
			fmt.Sprintf("@%s", segment),
			"D=D+M",
			fmt.Sprintf("@%d", index),
			"D=D+A",
		}, nil
	}
}

func PopCommand(target Command) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(PopStack),
			Constant("D=M"),
			target,
			Constant(
				"@SP",
				"A=M",
				"A=M",
				"A=D-A",
				"M=D-A",
			),
		)
	}
}

// PopStack sets the A register to point to the popped value in the stack.
func PopStack() ([]string, error) {
	return []string{
		"@SP",    // Load stack pointer segment
		"AM=M-1", // Decrement address of the stack pointer
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

// PushStack pushes the data currently present in D to the stack. It does not alter the data currently in D.
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
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant("D=M+D"),
		CommandFunc(PushStack),
	)
}

func SubCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant("D=M-D"),
		CommandFunc(PushStack),
	)
}

func NegCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant(
			"D=-M",
		),
		CommandFunc(PushStack),
	)
}

func AndCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant("D=M&D"),
		CommandFunc(PushStack),
	)
}

func OrCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant("D=M|D"),
		CommandFunc(PushStack),
	)
}

func NotCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant(
			"D=!M",
		),
		CommandFunc(PushStack),
	)
}

func EqCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant(
			"D=M-D",
			"@true",
			"D;JEQ",
			"D=0",
			"@end",
			"0;JMP",
			"(true)",
			"D=-1",
			"(end)",
		),
		CommandFunc(PushStack),
	)
}

func LtCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant(
			"D=M-D",
			"@true",
			"D;JLT",
			"D=0",
			"@end",
			"0;JMP",
			"(true)",
			"D=-1",
			"(end)",
		),
		CommandFunc(PushStack),
	)
}

func GtCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant(
			"D=M-D",
			"@true",
			"D;JGT",
			"D=0",
			"@end",
			"0;JMP",
			"(true)",
			"D=-1",
			"(end)",
		),
		CommandFunc(PushStack),
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
	token, err := l.Next()
	if err != nil {
		return nil, err
	}
	switch token.Variant {
	case push:
		return parsePushCommand(l)
	case pop:
		return parsePopCommand(l)
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
		return CommandFunc(EqCommand), nil
	case lt:
		return CommandFunc(LtCommand), nil
	case gt:
		return CommandFunc(GtCommand), nil
	default:
		panic(fmt.Errorf("variant %v is not yet supported", token.Variant))
	}
}

func parsePushCommand(l *lexer.Lexer[variant]) (Command, error) {
	src, err := l.Next()
	if err != nil {
		return nil, err
	}
	index, err := parseInteger(l)
	if err != nil {
		return nil, err
	}
	if src.Variant == constant {
		return PushCommand(ReadConstant(index)), nil
	}
	segment, ok := segments[src.Variant]
	if !ok {
		return nil, fmt.Errorf("unexpected push source %v", src.Literal)
	}
	return PushCommand(ReadSegment(segment, index)), nil
}

func parsePopCommand(l *lexer.Lexer[variant]) (Command, error) {
	target, err := l.Next()
	if err != nil {
		return nil, err
	}
	index, err := parseInteger(l)
	if err != nil {
		return nil, err
	}
	segment, ok := segments[target.Variant]
	if !ok {
		return nil, fmt.Errorf("unexpected pop target %v", target.Literal)
	}
	return PopCommand(SegmentTarget(segment, index)), nil
}

func parseInteger(l *lexer.Lexer[variant]) (int, error) {
	token, err := expect(l, integer)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(token.Literal)
}

func expect(l *lexer.Lexer[variant], v variant) (lexer.Token[variant], error) {
	token, err := l.Next()
	if err != nil {
		return lexer.Token[variant]{}, err
	}
	if token.Variant != v {
		return lexer.Token[variant]{}, fmt.Errorf("expected variant %v but found %v", v, token.Variant)
	}
	return token, nil
}
