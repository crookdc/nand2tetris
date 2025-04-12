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
		Constant("D=D+M"),
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
		Constant("D=D&M"),
		CommandFunc(PushStack),
	)
}

func OrCommand() ([]string, error) {
	return write(
		CommandFunc(PopStack),
		Constant("D=M"),
		CommandFunc(PopStack),
		Constant("D=D|M"),
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

func EqCommand(seq int32) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(PopStack),
			Constant("D=M"),
			CommandFunc(PopStack),
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
			CommandFunc(PushStack),
		)
	}
}

func LtCommand(seq int32) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(PopStack),
			Constant("D=M"),
			CommandFunc(PopStack),
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
			CommandFunc(PushStack),
		)
	}
}

func GtCommand(seq int32) CommandFunc {
	return func() ([]string, error) {
		return write(
			CommandFunc(PopStack),
			Constant("D=M"),
			CommandFunc(PopStack),
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
			CommandFunc(PushStack),
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
