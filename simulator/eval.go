package simulator

import (
	"errors"
	"github.com/crookdc/nand2tetris/hdl"
)

var (
	ErrInvalidSignal = errors.New("invalid signal")
	ErrInvalidID     = errors.New("invalid ECS id")
)

var NullWord = Word{}

type Signal struct {
	n int
}

func (s *Signal) Set(n int) error {
	if n != 1 && n != 0 {
		return ErrInvalidSignal
	}
	if s.n != n {
		s.n = n
	}
	return nil
}

type Word = [16]Signal

type ID = int

type ECS struct {
	free  []ID
	table []Word
	dirty []ID
}

func (e *ECS) Get(id ID) Word {
	return e.table[id]
}

func (e *ECS) Put(id ID, word Word) error {
	if id < 0 || id > len(e.table) {
		return ErrInvalidID
	}
	if word != e.table[id] && e.pristine(id) {
		e.dirty = append(e.dirty, id)
	}
	e.table[id] = word
	return nil
}

func (e *ECS) pristine(id ID) bool {
	for _, o := range e.dirty {
		if o == id {
			return false
		}
	}
	return true
}

func (e *ECS) Allocate() ID {
	if len(e.free) > 0 {
		id := e.free[len(e.free)-1]
		e.free = e.free[:len(e.free)-1]
		return id
	}
	e.table = append(e.table, Word{})
	return len(e.table) - 1
}

func (e *ECS) Free(id ID) error {
	if id < 0 || id > len(e.table) {
		return ErrInvalidID
	}
	e.table[id] = NullWord
	e.free = append(e.free, id)
	return nil
}

type Chip struct {
	Name    string
	Parent  ID
	ID      ID
	Inputs  []ID
	Outputs []ID
	Body    []hdl.Statement
}
