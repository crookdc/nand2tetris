package main

import (
	"flag"
	"fmt"
	"github.com/crookdc/nand2tetris/asm"
	"log"
	"os"
)

var (
	source = flag.String("source", "", "a file containing Hack assembly code")
)

func main() {
	flag.Parse()
	if *source == "" {
		log.Fatal("no source file provided")
	}
	src, err := os.ReadFile(*source)
	if err != nil {
		log.Fatal(err)
	}
	program, err := asm.Assemble(string(src))
	if err != nil {
		log.Fatal(err)
	}
	for _, ins := range program {
		var mc string
		for i := range 16 {
			mc += fmt.Sprintf("%v", ins[i])
		}
		fmt.Println(mc)
	}
}
