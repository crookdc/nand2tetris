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
				m: 15,
			},
			after: state{
				m:  15,
				pc: 15,
			},
		},
		{
			name:        "D;JGT",
			instruction: 0b111_0001100_000_001,
			before: state{
				d: 15,
			},
			after: state{
				d:  15,
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
