package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSolve(t *testing.T) {
	inputs := []struct {
		name  string
		board [][]byte
	}{
		{name: "default"},
		{name: "empty", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
		}},
		{name: "P1", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("P   "),
		}},
		{name: "P2", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte(" P  "),
		}},
		{name: "P3", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("  P "),
		}},
		{name: "P4", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("   P"),
		}},
		{name: "PX", board: [][]byte{
			[]byte("xxx "),
			[]byte(" P  "),
			[]byte("    "),
			[]byte("    "),
		}},
		{name: "R", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("R   "),
		}},
		{name: "B", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("B   "),
		}},
		{name: "K", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("K   "),
		}},
		{name: "NB", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("   x"),
			[]byte("  XN"),
		}},
		{name: "N", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("  XN"),
		}},
	}
	for _, in := range inputs {
		var out bytes.Buffer
		core := New(&out)
		if in.board != nil {
			core.board = in.board
		}
		core.Solve()
		if err := os.WriteFile(filepath.Join("testdata", in.name+".txt"), out.Bytes(), 0644); err != nil {
			t.Errorf("Solve %v got err %v", in, err)
		}
	}
}
