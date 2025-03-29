package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/crookdc/nand2tetris/hdl"
	"log"
	"os"
)

var (
	file   = flag.String("file", "", "name of file containing HDL under test")
	target = flag.String("target", "", "name of target under test")
	tests  = flag.String("tests", "", "name of test file")
)

func main() {
	flag.Parse()
	if *file == "" {
		log.Fatal("missing file name")
	}
	if *target == "" {
		log.Fatal("missing target name")
	}
	if *tests == "" {
		log.Fatal("missing comparison file name")
	}

	comparisons, err := loadTests(*tests)
	if err != nil {
		log.Fatal(err)
	}
	b := hdl.NewBreadboard()
	c, err := compileTarget(b)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range comparisons {
		if err := execute(t, c, b); err != nil {
			log.Fatal(err)
		}
	}
}

func execute(t test, c hdl.Chip, b *hdl.Breadboard) error {
	for name, value := range t.Inputs {
		err := b.SetGroup(c.Environment[name], binary(value))
		if err != nil {
			return err
		}
	}
	hdl.Tick(b)
	for i, expected := range t.Outputs {
		actual, err := b.GetGroup(c.Outputs[i])
		if err != nil {
			return err
		}
		if text(actual) != expected {
			return fmt.Errorf("expected output.%d to equal %s but got %s", i, expected, text(actual))
		}
	}
	return nil
}

func text(binary []byte) string {
	var n string
	for _, b := range binary {
		n += fmt.Sprintf("%d", b)
	}
	return n
}

func binary(text string) []byte {
	n := make([]byte, 0)
	for _, c := range text {
		if c == '1' {
			n = append(n, 1)
		} else {
			n = append(n, 0)
		}
	}
	return n
}

type test struct {
	Inputs  map[string]string `json:"inputs"`
	Outputs []string          `json:"outputs"`
}

func loadTests(filename string) ([]test, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var tests []test
	if err := json.Unmarshal(f, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

func compileTarget(b *hdl.Breadboard) (hdl.Chip, error) {
	support, err := hdl.ParseFile(*file)
	if err != nil {
		return hdl.Chip{}, err
	}
	compiled, err := hdl.Compile(b, support[*target], support)
	if err != nil {
		return hdl.Chip{}, err
	}
	return compiled, nil
}
