package lexer

func LineComment[T any](variant T, start string) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		previous := l.cursor
		for i := range start {
			if start[i] != c {
				// The tokens at the current position did not match the identifier for starting a line comment, so
				// rewind the cursor and return.
				l.cursor = previous
				return Token[T]{}, false, nil
			}
			l.cursor++
			c = l.source[l.cursor]
		}
		if err := l.Seek(Equals('\n')); err != nil {
			return Token[T]{}, false, err
		}
		l.cursor++
		if err := l.Seek(Not(l.ignored)); err != nil {
			return Token[T]{}, false, err
		}
		c = l.source[l.cursor]
		return Token[T]{
			Variant: variant,
		}, true, nil
	}
}

// Integer reads tokens that represent literal integers. The variant parameter defines what variant should be applied to
// an integer token when one has been processed. Integer does not care about signs and thus the input "-10" would not be
// considered an integer literal but rather some other token followed by an integer literal, in this case "10".
func Integer[T any](variant T) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if !Numerical(c) {
			return Token[T]{}, false, nil
		}
		token := Token[T]{
			Variant: variant,
			Literal: l.literal(Numerical),
		}
		return token, true, nil
	}
}

func String[T any](variant T) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if c != '"' {
			return Token[T]{}, false, nil
		}
		literal := l.literal(func(c uint8) bool {
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
type ConditionFunc func(uint8) bool

// Any constructs a complex ConditionFunc which yields true whenever any of the underlying ConditionFunc implementations
// do. The resulting ConditionFunc is effectively a sequence of the provided ConditionFunc implementations connected by
// an OR operator.
func Any(fns ...ConditionFunc) ConditionFunc {
	return func(c uint8) bool {
		for _, fn := range fns {
			if fn(c) {
				return true
			}
		}
		return false
	}
}

func All(fns ...ConditionFunc) ConditionFunc {
	return func(c uint8) bool {
		for _, fn := range fns {
			if !fn(c) {
				return false
			}
		}
		return true
	}
}

func Equals(target uint8) ConditionFunc {
	return func(c uint8) bool {
		return c == target
	}
}

func Not(fn ConditionFunc) ConditionFunc {
	return func(c uint8) bool {
		return !fn(c)
	}
}

// Condition takes a variant and a ConditionFunc and returns a Func capable of parsing bytes in a sequence that adhere to
// the conditions of the ConditionFunc.
func Condition[T any](variant T, fn ConditionFunc) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		if !fn(c) {
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
func Keywords[T any](keywords map[string]T, fn ConditionFunc) Func[T] {
	return func(l *Lexer[T], c uint8) (Token[T], bool, error) {
		previous := l.cursor
		if !fn(c) {
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
