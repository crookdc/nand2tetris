package main

import (
	"bufio"
	"flag"
	"github.com/crookdc/nand2tetris/simulator"
	"github.com/crookdc/nand2tetris/simulator/sdl"
	"log"
	"os"
)

var (
	rom = flag.String("rom", "", "path to file containing program that should be loaded into ROM")
)

func main() {
	flag.Parse()
	if *rom == "" {
		log.Fatal("mandatory rom flag not present")
	}
	program, err := parseProgram(*rom)
	if err != nil {
		log.Fatal(err)
	}
	s := simulator.New(simulator.Parameters{
		Screen:   simulator.Must(sdl.NewScreen()),
		Keyboard: sdl.NewKeyboard(),
		ROM:      program,
	})
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

func parseProgram(filename string) ([]uint16, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	program := make([]uint16, 0)
	scn := bufio.NewScanner(f)
	for scn.Scan() {
		text := scn.Text()
		var instruction uint16
		for i := range 16 {
			switch text[i] {
			case '0':
				instruction = instruction | uint16(0<<(15-i))
			case '1':
				instruction = instruction | uint16(1<<(15-i))
			default:
				panic("unexpected character in instruction " + text)
			}
		}
		program = append(program, instruction)
	}
	if err := scn.Err(); err != nil {
		return nil, err
	}
	return program, nil
}
