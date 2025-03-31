package simulator

import "fmt"

const (
	DestinationMaskA = 0b100
	DestinationMaskD = 0b010
	DestinationMaskM = 0b001

	jgt = 0b001
	jeq = 0b010
	jge = 0b011
	jlt = 0b100
	jne = 0b101
	jle = 0b110
	jmp = 0b111
)

var alu = map[uint8]func(*cpu) uint16{
	0b0101010: func(c *cpu) uint16 {
		return 0
	},
	0b0111111: func(c *cpu) uint16 {
		return 1
	},
	0b0111010: func(c *cpu) uint16 {
		return 0xFFFF
	},
	0b0001100: func(c *cpu) uint16 {
		return c.d
	},
	0b0110000: func(c *cpu) uint16 {
		return c.a
	},
	0b1110000: func(c *cpu) uint16 {
		return c.m
	},
	0b0001101: func(c *cpu) uint16 {
		return ^c.d
	},
	0b0110001: func(c *cpu) uint16 {
		return ^c.a
	},
	0b1110001: func(c *cpu) uint16 {
		return ^c.m
	},
	0b0001111: func(c *cpu) uint16 {
		return uint16(int16(c.d) * -1)
	},
	0b0110011: func(c *cpu) uint16 {
		return uint16(int16(c.a) * -1)
	},
	0b1110011: func(c *cpu) uint16 {
		return uint16(int16(c.m) * -1)
	},
	0b0011111: func(c *cpu) uint16 {
		return c.d + 1
	},
	0b0110111: func(c *cpu) uint16 {
		return c.a + 1
	},
	0b1110111: func(c *cpu) uint16 {
		return c.m + 1
	},
	0b0001110: func(c *cpu) uint16 {
		return c.d - 1
	},
	0b0110010: func(c *cpu) uint16 {
		return c.a - 1
	},
	0b1110010: func(c *cpu) uint16 {
		return c.m - 1
	},
	0b0000010: func(c *cpu) uint16 {
		return c.d + c.a
	},
	0b1000010: func(c *cpu) uint16 {
		return c.d + c.m
	},
	0b0010011: func(c *cpu) uint16 {
		return c.d - c.a
	},
	0b1010011: func(c *cpu) uint16 {
		return c.d - c.m
	},
	0b0000111: func(c *cpu) uint16 {
		return c.a - c.d
	},
	0b1000111: func(c *cpu) uint16 {
		return c.m - c.d
	},
	0b0000000: func(c *cpu) uint16 {
		return c.d & c.a
	},
	0b1000000: func(c *cpu) uint16 {
		return c.d & c.m
	},
	0b0010101: func(c *cpu) uint16 {
		return c.d | c.a
	},
	0b1010101: func(c *cpu) uint16 {
		return c.d | c.m
	},
}

type cpu struct {
	a  uint16
	d  uint16
	m  uint16
	pc uint16
}

func (c *cpu) address() uint16 {
	return c.a & 0b0111_1111_1111_1111
}

func (c *cpu) execute(in uint16) bool {
	if low(in, 16) {
		// The instruction is an A-instruction if the 16th bit is low. If not then it's a C-instruction.
		c.a = in
		c.pc++
		return false
	}
	return c.compute(in)
}

func (c *cpu) compute(instruction uint16) (w bool) {
	code := uint8((instruction >> 6) & 0b1111111)
	fn, ok := alu[code]
	if !ok {
		panic(fmt.Errorf("no computation registered for code %.7b at instruction %d", code, c.pc))
	}
	computed := fn(c)
	destination := (instruction >> 3) & 0b111
	if mask(destination, DestinationMaskA) {
		c.a = computed
	}
	if mask(destination, DestinationMaskD) {
		c.d = computed
	}
	if mask(destination, DestinationMaskM) {
		c.m = computed
		w = true
	}
	jump := false
	switch instruction & 0b111 {
	case jgt:
		jump = int(computed) > 0
	case jeq:
		jump = int(computed) == 0
	case jge:
		jump = int(computed) >= 0
	case jlt:
		jump = int(computed) < 0
	case jne:
		jump = int(computed) != 0
	case jle:
		jump = int(computed) <= 0
	case jmp:
		jump = true
	}
	if jump {
		c.pc = c.address()
	} else {
		c.pc++
	}
	return
}
