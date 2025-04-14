package main

import (
	"flag"
	"fmt"
	"github.com/crookdc/nand2tetris/vm"
	"log"
	"os"
)

var (
	file = flag.String("file", "", "a file containing vm code")
)

func main() {
	flag.Parse()
	if *file == "" {
		flag.Usage()
	}

	f, err := os.OpenFile(*file, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	asm, err := vm.Translate(*file, f)
	if err != nil {
		log.Fatal(err)
	}
	for _, ins := range asm {
		fmt.Println(ins)
	}
}
