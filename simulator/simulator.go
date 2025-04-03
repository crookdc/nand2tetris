package simulator

import (
	"time"
)

var (
	ScreenMemoryMapBegin     uint16 = 16_384
	KeyboardMemoryMapAddress        = 24_576
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

type Color struct {
	R uint8
	G uint8
	B uint8
}

type Screen interface {
	Clear() error
	Fill(color Color, points ...Point) error
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
	external := time.Tick(time.Second / 33)
	internal := time.Tick(time.Second / 300000)
	for {
		select {
		case _ = <-external:
			if err := s.draw(); err != nil {
				return err
			}
			s.ram[KeyboardMemoryMapAddress] = s.keyboard.Poll()
		case _ = <-internal:
			s.tick()
		default:
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
	white := make([]Point, 0)
	black := make([]Point, 0)
	for x := range 512 {
		for y := range 256 {
			point := Point{
				X: uint16(x),
				Y: uint16(y),
			}
			if low(s.ram[int(ScreenMemoryMapBegin)+y*32+x/16], x%16) {
				black = append(black, point)
			} else {
				white = append(white, point)
			}
		}
	}
	if err := s.screen.Fill(Color{255, 255, 255}, white...); err != nil {
		return err
	}
	if err := s.screen.Fill(Color{0, 0, 0}, black...); err != nil {
		return err
	}
	s.screen.Present()
	return nil
}

func high(word uint16, n int) bool {
	m := uint16(1)
	if n > 0 {
		m = m << n
	}
	return word&m != 0
}

func low(word uint16, n int) bool {
	return !high(word, n)
}

func mask(word uint16, m uint16) bool {
	return word&m == m
}
