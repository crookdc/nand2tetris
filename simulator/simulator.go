package simulator

import (
	"time"
)

var (
	ScreenMemoryMapBegin     uint16 = 16384
	ScreenMemoryMapLength    uint16 = 8192
	KeyboardMemoryMapAddress        = 24576
)

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type Point struct {
	X uint16
	Y uint16
}

type Screen interface {
	Clear() error
	Fill(points ...Point) error
	Present()
}

type Keyboard interface {
	Poll() uint16
}

type Parameters struct {
	Screen   Screen
	Keyboard Keyboard
	ROM      []uint16
}

func New(params Parameters) *Simulator {
	var rom [32768]uint16
	for i := range params.ROM {
		rom[i] = params.ROM[i]
	}
	return &Simulator{
		screen:   params.Screen,
		keyboard: params.Keyboard,
		rom:      rom,
		ram:      [32768]uint16{},
		cpu:      cpu{},
	}
}

type Simulator struct {
	screen   Screen
	keyboard Keyboard
	rom      [32768]uint16
	ram      [32768]uint16
	cpu      cpu
}

func (s *Simulator) Run() error {
	screen := time.Tick(time.Second / 30)
	for {
		select {
		case _ = <-screen:
			if err := s.draw(); err != nil {
				return err
			}
			screen = time.Tick(time.Second / 30)
		default:
			s.tick()
			s.ram[KeyboardMemoryMapAddress] = s.keyboard.Poll()
		}
	}
}

func (s *Simulator) tick() {
	instruction := s.rom[s.cpu.pc]
	s.cpu.m = s.ram[s.cpu.address()]
	if w := s.cpu.execute(instruction); w {
		s.ram[s.cpu.address()] = s.cpu.m
	}
}

func (s *Simulator) draw() error {
	points := make([]Point, 0)
	for i := range ScreenMemoryMapLength {
		pixels := s.ram[ScreenMemoryMapBegin+i]
		y := i / 32
		for j := range 16 {
			if low(pixels, j) {
				continue
			}
			x := ((i * 16) % 512) + uint16(j)
			points = append(points, Point{
				X: x,
				Y: y,
			})
		}
	}
	if err := s.screen.Clear(); err != nil {
		return err
	}
	if len(points) == 0 {
		return nil
	}
	if err := s.screen.Fill(points...); err != nil {
		return err
	}
	s.screen.Present()
	return nil
}

func high(word uint16, n int) bool {
	m := uint16(1)
	if n > 0 {
		m = m << (n - 1)
	}
	return word&m != 0
}

func low(word uint16, n int) bool {
	return !high(word, n)
}

func mask(word uint16, m uint16) bool {
	return word&m == m
}
