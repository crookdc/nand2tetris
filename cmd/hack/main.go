package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/crookdc/nand2tetris/internal/chip"
	"github.com/crookdc/nand2tetris/simulator"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
)

var (
	profile = flag.String("profile", "", "write profiling data to files with this base name")
	program = flag.String("program", "", "file containing program to be written to rom")
)

func main() {
	flag.Parse()
	if *profile != "" {
		f, err := os.Create(*profile + ".cpu")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *program == "" {
		log.Fatal("missing path to program")
	}

	rom, err := loadProgram(*program)
	if err != nil {
		log.Fatal(err)
	}
	sim, err := simulator.NewSDLSimulator(rom)
	if err != nil {
		log.Fatal(err)
	}
	defer sim.Close()
	for sim.Running {
		sim.Update()
	}

	if *profile != "" {
		f, err := os.Create(*profile + ".heap")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

func loadProgram(file string) (chip.ROM, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rom chip.ROM
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if len(line) != 16 {
			return nil, fmt.Errorf("invalid line length '%s'", line)
		}
		instruction, err := parseInstruction(line)
		if err != nil {
			return nil, err
		}
		rom = append(rom, instruction)
	}
	return rom, nil
}

func parseInstruction(line string) ([16]chip.Signal, error) {
	instruction := [16]chip.Signal{}
	for i := range 16 {
		bit, err := strconv.Atoi(string(line[i]))
		if err != nil {
			return [16]chip.Signal{}, err
		}
		instruction[i] = chip.Signal(bit)
	}
	return instruction, nil
}
