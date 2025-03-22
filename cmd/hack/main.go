package main

import (
	"flag"
	"github.com/crookdc/nand2tetris/hdl"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	hdlDir = flag.String("hdl-dir", "", "path to the directory containing HDL files")
)

func main() {
	flag.Parse()
	if *hdlDir == "" {
		log.Fatal("HDL directory must be provided")
	}
	chips, err := loadHdlDefinitions(*hdlDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(chips)
}

func loadHdlDefinitions(dir string) (map[string]hdl.ChipDefinition, error) {
	chips := make(map[string]hdl.ChipDefinition)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".hdl") {
			return nil
		}
		f, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parser := hdl.NewParser(hdl.Lexer{Source: string(f)})
		chs, err := parser.Parse()
		if err != nil {
			return err
		}
		for _, chip := range chs {
			chips[chip.Name] = chip
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chips, nil
}
