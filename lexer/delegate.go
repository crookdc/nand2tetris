package lexer

// LineComment produces a Func delegate that can be used to process line comments. A line comment is defined as a
// comment that starts with the character sequence provided in start and that stretches until the next linefeed
// character.
func LineComment[T comparable](start string) ConditionFunc[T] {
	return func(l *Lexer[T], c uint8) bool {
		previous := l.cursor
		for i := range start {
			if start[i] != c {
				// The tokens at the current position did not match the identifier for starting a line comment, so
				// rewind the cursor and return.
				l.cursor = previous
				return false
			}
			l.cursor++
			c = l.source[l.cursor]
		}
		if err := l.Seek(Equals[T]('\n')); err != nil {
			return false
		}
		return true
	}
}

// Integer reads tokens that represent literal integers. The variant parameter defines what variant should be applied to
// an integer token when one has been processed. Integer does not care about signs and thus the input "-10" would not be
// considered an integer literal but rather some other token followed by an integer literal, in this case "10".
func Integer[T comparable](variant T) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if !Numerical(l, c) {
			return Token[T]{}, false, nil
		}
		token := Token[T]{
			Variant: variant,
			Literal: l.literal(Numerical[T]),
		}
		return token, true, nil
	}
}

func StringLiteral[T comparable](variant T) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if c != '"' {
			return Token[T]{}, false, nil
		}
		l.cursor++
		literal := l.literal(func(lx *Lexer[T], c uint8) bool {
			return c != '"'
		})
		l.cursor++
		token := Token[T]{
			Variant: variant,
			Literal: literal,
		}
		return token, true, nil
	}
}

// ConditionFunc is a function that accepts a byte from the lexer and returns a boolean value based on whether the
// underlying condition is met. The lexer package provides several ConditionFunc values, such as Whitespace and
// Numerical.
type ConditionFunc[T comparable] func(lx *Lexer[T], c uint8) bool

// Any constructs a complex ConditionFunc which yields true whenever any of the underlying ConditionFunc implementations
// do. The resulting ConditionFunc is effectively a sequence of the provided ConditionFunc implementations connected by
// an OR operator.
func Any[T comparable](fns ...ConditionFunc[T]) ConditionFunc[T] {
	return func(lx *Lexer[T], c uint8) bool {
		for _, fn := range fns {
			if fn(lx, c) {
				return true
			}
		}
		return false
	}
}

// All products a composite ConditionFunc that returns true only when all the underlying ConditionFunc's do.
func All[T comparable](fns ...ConditionFunc[T]) ConditionFunc[T] {
	return func(lx *Lexer[T], c uint8) bool {
		for _, fn := range fns {
			if !fn(lx, c) {
				return false
			}
		}
		return true
	}
}

// Equals produces a ConditionFunc that returns true whenever c is equal to target and false otherwise.
func Equals[T comparable](target uint8) ConditionFunc[T] {
	return func(lx *Lexer[T], c uint8) bool {
		return c == target
	}
}

// Not produces a ConditionFunc that will always return the opposite of the result returned by fn.
func Not[T comparable](fn ConditionFunc[T]) ConditionFunc[T] {
	return func(lx *Lexer[T], c uint8) bool {
		return !fn(lx, c)
	}
}

// Condition takes a variant and a ConditionFunc and returns a Func capable of parsing bytes in a sequence that adhere to
// the conditions of the ConditionFunc.
func Condition[T comparable](variant T, fn ConditionFunc[T]) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if !fn(l, c) {
			return Token[T]{}, false, nil
		}
		token := Token[T]{
			Variant: variant,
			Literal: l.literal(fn),
		}
		return token, true, nil
	}
}

// Keywords constructs a Func capable of parsing any keyword in a map of keywords. It is the callers responsibility to
// provide a ConditionFunc that can properly cover all possible values of the keywords.
func Keywords[T comparable](keywords map[string]T, fn ConditionFunc[T]) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		previous := l.cursor
		if !fn(l, c) {
			return Token[T]{}, false, nil
		}
		literal := l.literal(fn)
		keyword, ok := keywords[literal]
		if !ok {
			l.cursor = previous
			return Token[T]{}, false, nil
		}
		token := Token[T]{
			Variant: keyword,
			Literal: literal,
		}
		return token, true, nil
	}
}
