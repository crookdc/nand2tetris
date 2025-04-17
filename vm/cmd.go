package vm

import "fmt"

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

func ReadNamed(lbl string) CommandFunc {
	return func() ([]string, error) {
		return []string{
			fmt.Sprintf("@%s", lbl),
			"D=A",
		}, nil
	}
}

func ReadNamedMemory(value string) CommandFunc {
	return func() ([]string, error) {
		return []string{
			fmt.Sprintf("@%s", value),
			"D=M",
		}, nil
	}
}

func ReadMemory(value int) CommandFunc {
	return func() ([]string, error) {
		if value < 0 || value > 32_767 {
			return nil, fmt.Errorf("invalid constant value %d", value)
		}
		return []string{
			fmt.Sprintf("@%d", value),
			"D=M",
		}, nil
	}
}

func ReadTemp(index int) CommandFunc {
	return func() ([]string, error) {
		if index < 0 || index > 7 {
			return nil, fmt.Errorf("invalid temp index %d", index)
		}
		index += 5
		return []string{
			fmt.Sprintf("@%d", index),
			"D=M",
		}, nil
	}
}

func ReadPointer(index int) CommandFunc {
	return func() ([]string, error) {
		var ptr string
		switch index {
		case 0:
			ptr = "@THIS"
		case 1:
			ptr = "@THAT"
		default:
			return nil, fmt.Errorf("invalid pointer index %d", index)
		}
		return []string{
			ptr,
			"D=M",
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
			CommandFunc(PushD),
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

func TempTarget(index int) CommandFunc {
	return func() ([]string, error) {
		if index < 0 || index > 7 {
			return nil, fmt.Errorf("invalid temp index %d", index)
		}
		// The temp segment is mapped directly to addresses 5-5+n where -1 < n < 8
		index += 5
		return []string{
			fmt.Sprintf("@%d", index),
			"D=D+A",
		}, nil
	}
}

func PointerTarget(index int) CommandFunc {
	return func() ([]string, error) {
		var ptr string
		switch index {
		case 0:
			ptr = "@THIS"
		case 1:
			ptr = "@THAT"
		default:
			return nil, fmt.Errorf("invalid pointer index %d", index)
		}
		return []string{
			ptr,
			"D=D+A",
		}, nil
	}
}

func MemoryTarget(address int) CommandFunc {
	return func() ([]string, error) {
		return []string{
			fmt.Sprintf("@%d", address),
			"D=D+A",
		}, nil
	}
}

func PopCommand(target Command) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(Pop),
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

// Pop sets the A register to point to the popped value in the stack.
func Pop() ([]string, error) {
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

// PushD pushes the data currently present in D to the stack. It does not alter the data currently in D.
func PushD() ([]string, error) {
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
		CommandFunc(Pop),
		Constant("D=M"),
		CommandFunc(Pop),
		Constant("D=D+M"),
		CommandFunc(PushD),
	)
}

func SubCommand() ([]string, error) {
	return write(
		CommandFunc(Pop),
		Constant("D=M"),
		CommandFunc(Pop),
		Constant("D=M-D"),
		CommandFunc(PushD),
	)
}

func NegCommand() ([]string, error) {
	return write(
		CommandFunc(Pop),
		Constant(
			"D=-M",
		),
		CommandFunc(PushD),
	)
}

func AndCommand() ([]string, error) {
	return write(
		CommandFunc(Pop),
		Constant("D=M"),
		CommandFunc(Pop),
		Constant("D=D&M"),
		CommandFunc(PushD),
	)
}

func OrCommand() ([]string, error) {
	return write(
		CommandFunc(Pop),
		Constant("D=M"),
		CommandFunc(Pop),
		Constant("D=D|M"),
		CommandFunc(PushD),
	)
}

func NotCommand() ([]string, error) {
	return write(
		CommandFunc(Pop),
		Constant(
			"D=!M",
		),
		CommandFunc(PushD),
	)
}

func EqCommand(seq int) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(Pop),
			Constant("D=M"),
			CommandFunc(Pop),
			Constant(
				"D=M-D",
				fmt.Sprintf("@true_%d", seq),
				"D;JEQ",
				"D=0",
				fmt.Sprintf("@end_%d", seq),
				"0;JMP",
				fmt.Sprintf("(true_%d)", seq),
				"D=-1",
				fmt.Sprintf("(end_%d)", seq),
			),
			CommandFunc(PushD),
		)
	}
}

func LtCommand(seq int) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(Pop),
			Constant("D=M"),
			CommandFunc(Pop),
			Constant(
				"D=M-D",
				fmt.Sprintf("@true_%d", seq),
				"D;JLT",
				"D=0",
				fmt.Sprintf("@end_%d", seq),
				"0;JMP",
				fmt.Sprintf("(true_%d)", seq),
				"D=-1",
				fmt.Sprintf("(end_%d)", seq),
			),
			CommandFunc(PushD),
		)
	}
}

func GtCommand(seq int) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(Pop),
			Constant("D=M"),
			CommandFunc(Pop),
			Constant(
				"D=M-D",
				fmt.Sprintf("@true_%d", seq),
				"D;JGT",
				"D=0",
				fmt.Sprintf("@end_%d", seq),
				"0;JMP",
				fmt.Sprintf("(true_%d)", seq),
				"D=-1",
				fmt.Sprintf("(end_%d)", seq),
			),
			CommandFunc(PushD),
		)
	}
}

func FunctionCommand(name string, vars int) CommandFunc {
	return func() ([]string, error) {
		commands := make([]Command, 0, vars+1)
		commands = append(commands, Constant(fmt.Sprintf("(%s)", name)))
		for range vars {
			commands = append(commands, PushCommand(ReadConstant(0)))
		}
		return write(commands...)
	}
}

func ReturnCommand() ([]string, error) {
	return write(
		Constant(
			"@LCL",
			"D=M",
			"@R13",
			"M=D",
			"@5",
			"D=D-A",
			// return address
			"@R14",
			"M=D",
			// ARG=pop
			"@SP",
			"AM=M-1",
			"D=M",
			"@ARG",
			"A=M",
			"M=D",
			"@ARG",
			"D=M",
			"@SP",
			"M=D+1",
			// THAT=frame-1
			"@R13",
			"A=M-1",
			"D=M",
			"@THAT",
			"M=D",
			// THIS=frame-2
			"@R13",
			"D=M",
			"@2",
			"D=D-A",
			"A=D",
			"D=M",
			"@THIS",
			"M=D",
			// ARG=frame-3
			"@R13",
			"D=M",
			"@3",
			"D=D-A",
			"A=D",
			"D=M",
			"@ARG",
			"M=D",
			// LCL=frame-4
			"@R13",
			"D=M",
			"@4",
			"D=D-A",
			"A=D",
			"D=M",
			"@LCL",
			"M=D",
			// JMP to return address
			"@R14",
			"0;JMP",
		),
	)
}

func Subtract(loader Command, n int) CommandFunc {
	if n < 0 {
		panic(fmt.Errorf("invalid subtraction operand: %v", n))
	}
	return func() ([]string, error) {
		asm, err := loader.Evaluate()
		if err != nil {
			return nil, err
		}
		asm = append(
			asm,
			fmt.Sprintf("@%d", n),
			"D=D-A",
		)
		return asm, nil
	}
}

func CallCommand(caller string, callee string, n int, seq int) CommandFunc {
	return func() ([]string, error) {
		return write(
			PushCommand(ReadNamed(fmt.Sprintf("%s$ret%d", caller, seq))),
			PushCommand(ReadNamedMemory("LCL")),
			PushCommand(ReadNamedMemory("ARG")),
			PushCommand(ReadNamedMemory("THIS")),
			PushCommand(ReadNamedMemory("THAT")),
			Constant(
				"@SP",
				"D=M",
				"@LCL",
				"M=D",
			),
			Subtract(ReadNamedMemory("SP"), 5+n),
			Constant(
				"@ARG",
				"M=D",
			),
			Constant(
				fmt.Sprintf("@%s", callee),
				"0;JMP",
				fmt.Sprintf("(%s.ret$%d)", caller, seq),
			),
		)
	}
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
