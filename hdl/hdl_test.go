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
	if testing.Short() {
		t.Skip("skipping HDL tests in short mode")
	}

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

	type test struct {
		inputs   map[string][]byte
		expected [][]byte
	}
	tests := []struct {
		chip  string
		tests []test
	}{
		{
			chip: "mux_2",
			tests: []test{
				{
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
					inputs: map[string][]byte{
						"s": {1},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
					},
					expected: [][]byte{
						{1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
					},
				},
			},
		},
		{
			chip: "mux_4",
			tests: []test{
				{
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
			},
		},
		{
			chip: "mux_8",
			tests: []test{
				{
					inputs: map[string][]byte{
						"s": {0, 0, 0},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {0, 0, 1},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {0, 1, 0},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {0, 1, 1},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {1, 0, 0},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {1, 0, 1},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {1, 1, 0},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s": {1, 1, 1},
						"a": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1},
						"c": {0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0},
						"d": {1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1},
						"e": {1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"f": {1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 1},
						"g": {0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0},
						"h": {1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1},
					},
				},
			},
		},
		{
			chip: "dmux_2",
			tests: []test{
				{
					inputs: map[string][]byte{
						"s":  {0},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s":  {1},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
			},
		},
		{
			chip: "dmux_4",
			tests: []test{
				{
					inputs: map[string][]byte{
						"s":  {0, 0},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s":  {0, 1},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s":  {1, 0},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"s":  {1, 1},
						"in": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
			},
		},
		{
			chip: "half_adder",
			tests: []test{
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {0},
					},
					expected: [][]byte{
						{0},
						{0},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {1},
					},
					expected: [][]byte{
						{0},
						{1},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {1},
						"b": {0},
					},
					expected: [][]byte{
						{0},
						{1},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {1},
						"b": {1},
					},
					expected: [][]byte{
						{1},
						{0},
					},
				},
			},
		},
		{
			chip: "full_adder",
			tests: []test{
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {0},
						"c": {0},
					},
					expected: [][]byte{
						{0},
						{0},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {0},
						"c": {1},
					},
					expected: [][]byte{
						{0},
						{1},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {1},
						"c": {0},
					},
					expected: [][]byte{
						{0},
						{1},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {1},
						"b": {0},
						"c": {0},
					},
					expected: [][]byte{
						{0},
						{1},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {0},
						"b": {1},
						"c": {1},
					},
					expected: [][]byte{
						{1},
						{0},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {1},
						"b": {1},
						"c": {1},
					},
					expected: [][]byte{
						{1},
						{1},
					},
				},
			},
		},
		{
			chip: "adder_16",
			tests: []test{
				{
					inputs: map[string][]byte{
						"a": {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"b": {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"a": {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
						"b": {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
			},
		},
		{
			chip: "alu",
			tests: []test{
				{
					// 0
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {0},
						"zy": {1},
						"ny": {0},
						"f":  {1},
						"n":  {0},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{1},
						{0},
					},
				},
				{
					// -1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {1},
						"ny": {0},
						"f":  {1},
						"n":  {0},
					},
					expected: [][]byte{
						{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
						{0},
						{1},
					},
				},
				{
					// 1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {1},
						"ny": {1},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
						{0},
						{0},
					},
				},
				{
					// x + y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {0},
						"zy": {0},
						"ny": {0},
						"f":  {1},
						"n":  {0},
					},
					expected: [][]byte{
						{1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0},
						{0},
						{1},
					},
				},
				{
					// x + 1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {1},
						"zy": {1},
						"ny": {1},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0},
						{0},
						{0},
					},
				},
				{
					// x - 1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {0},
						"zy": {1},
						"ny": {1},
						"f":  {1},
						"n":  {0},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						{0},
						{0},
					},
				},
				{
					// y - 1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {0},
						"ny": {0},
						"f":  {1},
						"n":  {0},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						{0},
						{1},
					},
				},
				{
					// y + 1
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {0},
						"ny": {1},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0},
						{0},
						{0},
					},
				},
				{
					// x - y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"zx": {0},
						"nx": {1},
						"zy": {0},
						"ny": {0},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
						{0},
						{0},
					},
				},
				{
					// x - y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {0},
						"zy": {0},
						"ny": {1},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
						{0},
						{0},
					},
				},
				{
					// -x
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {0},
						"zy": {1},
						"ny": {1},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0},
						{0},
						{1},
					},
				},
				{
					// -y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"zx": {1},
						"nx": {1},
						"zy": {0},
						"ny": {0},
						"f":  {1},
						"n":  {1},
					},
					expected: [][]byte{
						{1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0},
						{0},
						{1},
					},
				},
				{
					// x & y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"zx": {0},
						"nx": {0},
						"zy": {0},
						"ny": {0},
						"f":  {0},
						"n":  {0},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						{0},
						{0},
					},
				},
				{
					// x | y
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"zx": {0},
						"nx": {1},
						"zy": {0},
						"ny": {1},
						"f":  {0},
						"n":  {1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0},
						{1},
					},
				},
				{
					// !x
					inputs: map[string][]byte{
						"x":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"y":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"zx": {0},
						"nx": {0},
						"zy": {1},
						"ny": {1},
						"f":  {0},
						"n":  {1},
					},
					expected: [][]byte{
						{1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 1, 0, 0, 1, 0, 0},
						{0},
						{1},
					},
				},
				{
					// !y
					inputs: map[string][]byte{
						"x":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {0},
						"ny": {0},
						"f":  {0},
						"n":  {1},
					},
					expected: [][]byte{
						{1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 1, 0, 0, 1, 0, 0},
						{0},
						{1},
					},
				},
				{
					// x
					inputs: map[string][]byte{
						"x":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {0},
						"nx": {0},
						"zy": {1},
						"ny": {1},
						"f":  {0},
						"n":  {0},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						{0},
						{1},
					},
				},
				{
					// y
					inputs: map[string][]byte{
						"x":  {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0},
						"y":  {0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						"zx": {1},
						"nx": {1},
						"zy": {0},
						"ny": {0},
						"f":  {0},
						"n":  {0},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
						{0},
						{0},
					},
				},
			},
		},
	}

	for _, group := range tests {
		t.Run(group.chip, func(t *testing.T) {
			for _, test := range group.tests {
				evaluator := hdl.NewEvaluator(definitions)
				chip, err := evaluator.Evaluate(definitions[group.chip])
				if err != nil {
					t.Fatal(err)
				}
				for arg, value := range test.inputs {
					if err := evaluator.Breadboard.SetGroup(chip.Environment[arg], value); err != nil {
						t.Fatal(err)
					}
				}
				hdl.Tick(evaluator.Breadboard)
				for i, expected := range test.expected {
					actual, err := evaluator.Breadboard.GetGroup(chip.Outputs[i])
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(actual, expected) {
						t.Errorf("expected %v but got %v on output %d", expected, actual, i)
					}
				}
			}
		})
	}
}

func TestProgramCounter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HDL tests in short mode")
	}

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

	type test struct {
		inputs   map[string][]byte
		expected [][]byte
	}
	tests := []struct {
		tests []test
	}{
		{
			tests: []test{
				{
					inputs: map[string][]byte{
						"load": {1},
						"inc":  {0},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"load": {0},
						"inc":  {0},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"load": {0},
						"inc":  {1},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"load": {0},
						"inc":  {0},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0},
					},
				},
			},
		},
		{
			tests: []test{
				{
					inputs: map[string][]byte{
						"load": {1},
						"inc":  {0},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{
					inputs: map[string][]byte{
						"load": {0},
						"inc":  {0},
						"rst":  {1},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
				},
				{
					inputs: map[string][]byte{
						"load": {1},
						"inc":  {0},
						"rst":  {0},
						"in":   {1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1},
					},
					expected: [][]byte{
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
			},
		},
	}
	for _, test := range tests {
		evaluator := hdl.NewEvaluator(definitions)
		chip, err := evaluator.Evaluate(definitions["program_counter"])
		if err != nil {
			t.Fatal(err)
		}
		for _, step := range test.tests {
			for arg, value := range step.inputs {
				if err := evaluator.Breadboard.SetGroup(chip.Environment[arg], value); err != nil {
					t.Fatal(err)
				}
			}
			hdl.Tick(evaluator.Breadboard)
			for i, expected := range step.expected {
				actual, err := evaluator.Breadboard.GetGroup(chip.Outputs[i])
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("expected %v but got %v on output %d", expected, actual, i)
				}
			}
		}
	}
}

func TestCPU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HDL tests in short mode")
	}

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

	evaluator := hdl.NewEvaluator(definitions)
	chip, err := evaluator.Evaluate(definitions["cpu"])
	if err != nil {
		t.Fatal(err)
	}

	hdl.Tick(evaluator.Breadboard)
	value := []byte{0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 1}
	err = evaluator.Breadboard.SetGroup(chip.Environment["instruction"], value)
	if err != nil {
		t.Fatal(err)
	}
	hdl.Tick(evaluator.Breadboard)
	err = evaluator.Breadboard.SetGroup(chip.Environment["instruction"], []byte{1, 1, 1, 0, 1, 1, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	hdl.Tick(evaluator.Breadboard)

	output, err := evaluator.Breadboard.GetGroup(chip.Outputs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(output, value) {
		t.Errorf("expected %v but got %v", value, output)
	}
}
