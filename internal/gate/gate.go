// Package gate provides implementations of basic logical gates in three different flavours which accept gate uint8 and
// uint16 parameters respectively. All gates are built on top of the NotAnd gate.
package gate

func NotAnd(a, b uint8) uint8 {
	res := uint8(0)
	for i := range 8 {
		if (a>>i)&1 == 0 && (b>>i)&1 == 0 {
			res = res | 1<<i
		}
	}
	return res
}

func NotAndUint16(a, b uint16) uint16 {
	am, al := splitUint16(a)
	bm, bl := splitUint16(b)
	return joinUint16(NotAnd(am, bm), NotAnd(al, bl))
}

func Not(a uint8) uint8 {
	return NotAnd(a, 0)
}

func NotUint16(a uint16) uint16 {
	msb, lsb := splitUint16(a)
	return joinUint16(Not(msb), Not(lsb))
}

func And(a, b uint8) uint8 {
	return NotAnd(Not(a), Not(b))
}

func AndUint16(a, b uint16) uint16 {
	am, al := splitUint16(a)
	bm, bl := splitUint16(b)
	return joinUint16(And(am, bm), And(al, bl))
}

func Or(a, b uint8) uint8 {
	return Not(NotAnd(a, b))
}

func OrUint16(a, b uint16) uint16 {
	am, al := splitUint16(a)
	bm, bl := splitUint16(b)
	return joinUint16(Or(am, bm), Or(al, bl))
}

func Xor(a, b uint8) uint8 {
	return Or(And(a, Not(b)), And(Not(a), b))
}

func XorUint16(a, b uint16) uint16 {
	am, al := splitUint16(a)
	bm, bl := splitUint16(b)
	return joinUint16(Xor(am, bm), Xor(al, bl))
}

// Mux2Way provides a multiplexer for 2 inputs and a selector. This variant of the multiplexer supports
// only binary values (0, 1) to be passed in as selector, any non-zero value is considered set (0xFF) and
// only zero is considered unset (0x00). The multiplexer will return the value of `a` if `s` is unset (0)
// and the value of `b` is `s` is set (> 0).
func Mux2Way(s uint8, a, b uint16) uint16 {
	s = selector(s)
	return OrUint16(AndUint16(NotUint16(uint16(s)|uint16(s)<<8), a), AndUint16(uint16(s)|uint16(s)<<8, b))
}

// Mux4Way provides a multiplexer for 4 inputs and a selector consisting of 2 bytes. Non-zero values on
// selector bytes are considered as set and only zero is considered unset.
func Mux4Way(s [2]uint8, a, b, c, d uint16) uint16 {
	ab := Mux2Way(s[1], a, b)
	cd := Mux2Way(s[1], c, d)
	return Mux2Way(s[0], ab, cd)
}

// Mux8Way provides a multiplexer for 8 inputs and a selector consisting of 3 bytes. Non-zero values on
// selector bytes are considered as set and only zero is considered unset.
func Mux8Way(s [3]uint8, a, b, c, d, e, f, g, h uint16) uint16 {
	abcd := Mux4Way([2]uint8{s[1], s[2]}, a, b, c, d)
	efgh := Mux4Way([2]uint8{s[1], s[2]}, e, f, g, h)
	return Mux2Way(s[0], abcd, efgh)
}

func selector(n uint8) uint8 {
	if n > 0 {
		return 0xFF
	}
	return 0
}

func splitUint16(n uint16) (msb uint8, lsb uint8) {
	msb = uint8(n >> 8)
	lsb = uint8(n)
	return msb, lsb
}

func joinUint16(msb, lsb uint8) uint16 {
	return uint16(msb)<<8 | uint16(lsb)
}
