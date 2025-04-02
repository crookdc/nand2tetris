package simulator

import (
	"testing"
)

func TestCPU_compute(t *testing.T) {
	type state struct {
		a  uint16
		d  uint16
		m  uint16
		pc uint16
	}
	tests := []struct {
		name        string
		instruction uint16
		before      state

		after state
		write bool
	}{
		{
			name:        "M=A",
			instruction: 0b111_0110000_001_000,
			before: state{
				a: 1235,
			},
			after: state{
				a:  1235,
				m:  1235,
				pc: 1,
			},
			write: true,
		},
		{
			name:        "D=A+1",
			instruction: 0b111_0110111_010_000,
			before: state{
				a: 1235,
			},
			after: state{
				a:  1235,
				d:  1236,
				pc: 1,
			},
		},
		{
			name:        "@16384",
			instruction: 0b010_0000000_000_000,
			before: state{
				a: 0,
			},
			after: state{
				a:  16384,
				pc: 1,
			},
		},
		{
			name:        "M;JGT",
			instruction: 0b111_1110000_000_001,
			before: state{
				a: 15,
				m: 10,
			},
			after: state{
				a:  15,
				m:  10,
				pc: 15,
			},
		},
		{
			name:        "D;JGT",
			instruction: 0b111_0001100_000_001,
			before: state{
				a: 15,
				d: 23,
			},
			after: state{
				a:  15,
				d:  23,
				pc: 15,
			},
		},
		{
			name:        "D;JGT",
			instruction: 0b111_0001100_000_001,
			before: state{
				d: 0,
			},
			after: state{
				d:  0,
				pc: 1,
			},
		},
		{
			name:        "D=-D",
			instruction: 0b111_0001111_010_000,
			before: state{
				d: 76,
			},
			after: state{
				d:  65460,
				pc: 1,
			},
		},
		{
			name:        "A=-A",
			instruction: 0b111_0110011_100_000,
			before: state{
				a: 76,
			},
			after: state{
				a:  65460,
				pc: 1,
			},
		},
		{
			name:        "M=-M",
			instruction: 0b111_1110011_001_000,
			before: state{
				m: 76,
			},
			after: state{
				m:  65460,
				pc: 1,
			},
			write: true,
		},
		{
			name:        "AM=!D",
			instruction: 0b111_0001101_101_000,
			before: state{
				d: 0b1110_0000_1011_1110,
			},
			after: state{
				a:  0b0001_1111_0100_0001,
				d:  0b1110_0000_1011_1110,
				m:  0b0001_1111_0100_0001,
				pc: 1,
			},
			write: true,
		},
		{
			name:        "D=A-D",
			instruction: 0b111_0000111_010_000,
			before: state{
				a: 20,
				d: 1,
			},
			after: state{
				a:  20,
				d:  19,
				pc: 1,
			},
		},
		{
			name:        "@24576",
			instruction: 0b0110_0000_0000_0000,
			before: state{
				a: 20,
				d: 1,
			},
			after: state{
				a:  24576,
				d:  1,
				pc: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := cpu{
				a:  test.before.a,
				d:  test.before.d,
				m:  test.before.m,
				pc: test.before.pc,
			}
			write := c.execute(test.instruction)
			if c.a != test.after.a {
				t.Errorf("expected a to equal %v but got %v", test.after.a, c.a)
			}
			if c.d != test.after.d {
				t.Errorf("expected d to equal %v but got %v", test.after.d, c.d)
			}
			if c.m != test.after.m {
				t.Errorf("expected m to equal %v but got %v", test.after.m, c.m)
			}
			if c.pc != test.after.pc {
				t.Errorf("expected pc to equal %v but got %v", test.after.pc, c.pc)
			}
			if write != test.write {
				t.Errorf("expected write to equal %v but got %v", test.write, write)
			}
		})
	}
}
