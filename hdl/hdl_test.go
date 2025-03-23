package hdl_test

import (
	"github.com/crookdc/nand2tetris/hdl"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestHDL(t *testing.T) {
	definitions := make(map[string]hdl.ChipDefinition)
	err := filepath.WalkDir("../bins/hdl", func(path string, d fs.DirEntry, err error) error {
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
			definitions[chip.Name] = chip
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		chip     string
		inputs   map[string][]byte
		expected [][]byte
	}{
		{
			chip: "mux_2",
			inputs: map[string][]byte{
				"s": {0},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
			},
			expected: [][]byte{
				{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
			},
		},
		{
			chip: "mux_2",
			inputs: map[string][]byte{
				"s": {1},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
			},
			expected: [][]byte{
				{1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
			},
		},
		{
			chip: "mux_4",
			inputs: map[string][]byte{
				"s": {0, 0},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
				"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
				"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
			},
			expected: [][]byte{
				{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
			},
		},
		{
			chip: "mux_4",
			inputs: map[string][]byte{
				"s": {0, 1},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
				"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
				"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
			},
			expected: [][]byte{
				{1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
			},
		},
		{
			chip: "mux_4",
			inputs: map[string][]byte{
				"s": {1, 0},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
				"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
				"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
			},
			expected: [][]byte{
				{0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
			},
		},
		{
			chip: "mux_4",
			inputs: map[string][]byte{
				"s": {1, 1},
				"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
				"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
				"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
				"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
			},
			expected: [][]byte{
				{1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
			},
		},
	}

	for _, test := range tests {
		compiler := hdl.NewCompiler(definitions)
		chip, err := compiler.Compile(definitions[test.chip])
		if err != nil {
			t.Fatal(err)
		}
		for arg, value := range test.inputs {
			if err := compiler.Breadboard.SetGroup(chip.Inputs[arg], value); err != nil {
				t.Fatal(err)
			}
		}
		for i, expected := range test.expected {
			actual, err := compiler.Breadboard.GetGroup(chip.Outputs[i])
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("expected %v but got %v on output %d", expected, actual, i)
			}
		}
	}
}
