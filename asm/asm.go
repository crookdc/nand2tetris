package asm

import (
	"fmt"
	"strconv"
)

var (
	computations = map[string]int{
		"0":   0b0101010,
		"1":   0b0111111,
		"-1":  0b0111010,
		"D":   0b0001100,
		"A":   0b0110000,
		"M":   0b1110000,
		"!D":  0b0001101,
		"!A":  0b0110001,
		"!M":  0b1110001,
		"-D":  0b0001111,
		"-A":  0b0110011,
		"-M":  0b1110011,
		"D+1": 0b0011111,
		"A+1": 0b0110111,
		"M+1": 0b1110111,
		"D-1": 0b0001110,
		"A-1": 0b0110010,
		"M-1": 0b1110010,
		"D+A": 0b0000010,
		"D+M": 0b1000010,
		"D-A": 0b0010011,
		"D-M": 0b1010011,
		"A-D": 0b0000111,
		"M-D": 0b1000111,
		"D&A": 0b0000000,
		"D&M": 0b1000000,
		"D|A": 0b0010101,
		"D|M": 0b1010101,
	}
	jumps = map[string]int{
		"JGT": 0b001,
		"JEQ": 0b010,
		"JGE": 0b011,
		"JLT": 0b100,
		"JNE": 0b101,
		"JLE": 0b110,
		"JMP": 0b111,
	}
	destinations = map[uint8]int{
		'M': 0b001,
		'D': 0b010,
		'A': 0b100,
	}
)

func Assemble(src string) ([][16]byte, error) {
	mem, err := buildMemoryMap(src)
	if err != nil {
		return nil, err
	}
	var program [][16]byte
	ps := parser{
		lexer: lexer{
			src: src,
		},
	}
	for ps.more() {
		ins, err := ps.next()
		if err != nil {
			return nil, err
		}
		switch v := ins.(type) {
		case load:
			bin, err := assembleLoadInstruction(mem, v)
			if err != nil {
				return nil, err
			}
			program = append(program, bin)
		case compute:
			bin, err := assembleComputeInstruction(v)
			if err != nil {
				return nil, err
			}
			program = append(program, bin)
		default:
		}
	}
	return program, nil
}

func wrap(n int) [16]byte {
	var r [16]byte
	for i := range 16 {
		r[i] = byte(n>>(15-i)) & 1
	}
	return r
}

func assembleLoadInstruction(mem map[string]int, v load) ([16]byte, error) {
	var bin int
	var err error
	if v.value.variant == integer {
		bin, err = strconv.Atoi(v.value.literal)
	} else if v.value.variant == identifier {
		bin = mem[v.value.literal]
	} else {
		return [16]byte{}, fmt.Errorf("unexpected load token %+v", v.value)
	}
	if err != nil {
		return [16]byte{}, err
	}
	bin = bin & 0b0111_1111_1111_1111
	return wrap(bin), nil
}

func assembleComputeInstruction(v compute) ([16]byte, error) {
	bin, ok := computations[v.comp]
	if !ok {
		return [16]byte{}, fmt.Errorf("unexpected computational segment %s", v.comp)
	}
	bin = bin << 6
	if v.dest != nil {
		dest := 0
		for i := range v.dest.literal {
			d, ok := destinations[v.dest.literal[i]]
			if !ok {
				return [16]byte{}, fmt.Errorf("invalid destination %+v", v.dest)
			}
			dest = dest | d
		}
		bin = bin | (dest << 3)
	}
	if v.jump != nil {
		jump, ok := jumps[v.jump.literal]
		if !ok {
			return [16]byte{}, fmt.Errorf("invalid jump %+v", v.jump)
		}
		bin = bin | jump
	}
	bin = bin | 0b1110_0000_0000_0000
	return wrap(bin), nil
}

func buildMemoryMap(src string) (map[string]int, error) {
	mem := map[string]int{
		"R0":     0,
		"R1":     1,
		"R2":     2,
		"R3":     3,
		"R4":     4,
		"R5":     5,
		"R6":     6,
		"R7":     7,
		"R8":     8,
		"R9":     9,
		"R10":    10,
		"R11":    11,
		"R12":    12,
		"R13":    13,
		"R14":    14,
		"R15":    15,
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
		"SCREEN": 16_384,
		"KBD":    24_576,
	}
	ps := parser{
		lexer: lexer{
			src: src,
		},
	}
	for line := 0; ps.more(); line++ {
		ins, err := ps.next()
		if err != nil {
			return nil, err
		}
		switch v := ins.(type) {
		case label:
			if _, ok := mem[v.value.literal]; !ok {
				mem[v.value.literal] = line
			}
			line--
		default:
		}
	}
	ps = parser{
		lexer: lexer{
			src: src,
		},
	}
	cursor := 16
	for ps.more() {
		ins, err := ps.next()
		if err != nil {
			return nil, err
		}
		switch v := ins.(type) {
		case load:
			if v.value.variant != identifier {
				continue
			}
			_, ok := mem[v.value.literal]
			if !ok {
				mem[v.value.literal] = cursor
				cursor++
			}
		default:
		}
	}
	return mem, nil
}
