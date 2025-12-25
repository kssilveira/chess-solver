package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSolve(t *testing.T) {
	inputs := []struct {
		name          string
		board         [][]byte
		maxDepth      int
		maxPrintDepth int
	}{
		{name: "default", maxDepth: 2},
		{name: "empty", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
		}},
		{name: "P1", board: [][]byte{
			[]byte("   p"),
			[]byte("    "),
			[]byte("    "),
			[]byte("P   "),
		}},
		{name: "P2", board: [][]byte{
			[]byte("  p "),
			[]byte("    "),
			[]byte("    "),
			[]byte(" P  "),
		}},
		{name: "P3", board: [][]byte{
			[]byte(" p  "),
			[]byte("    "),
			[]byte("    "),
			[]byte("  P "),
		}},
		{name: "P4", board: [][]byte{
			[]byte("p   "),
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
			[]byte("   r"),
			[]byte("    "),
			[]byte("    "),
			[]byte("R   "),
		}},
		{name: "B", board: [][]byte{
			[]byte("   b"),
			[]byte("    "),
			[]byte("    "),
			[]byte("B   "),
		}},
		{name: "K", board: [][]byte{
			[]byte("   k"),
			[]byte("    "),
			[]byte("    "),
			[]byte("K   "),
		}},
		{name: "Kk", board: [][]byte{
			[]byte("    "),
			[]byte("  k "),
			[]byte(" K  "),
			[]byte("    "),
		}},
		{name: "Kk2", board: [][]byte{
			[]byte("    "),
			[]byte(" k  "),
			[]byte("    "),
			[]byte("K k "),
		}},
		{name: "NB", board: [][]byte{
			[]byte("nx  "),
			[]byte("X   "),
			[]byte("   x"),
			[]byte("  XN"),
		}},
		{name: "N", board: [][]byte{
			[]byte("nx  "),
			[]byte("    "),
			[]byte("    "),
			[]byte("  XN"),
		}},
		{name: "Nk", maxDepth: 3, board: [][]byte{
			[]byte("k R "),
			[]byte("    "),
			[]byte("RR  "),
			[]byte("   N"),
		}},
		{name: "RNk", maxDepth: 500000, maxPrintDepth: 5, board: [][]byte{
			[]byte("  R "),
			[]byte("k   "),
			[]byte(" R  "),
			[]byte("R  N"),
		}},
	}
	for _, in := range inputs {
		config := Config{MaxDepth: 5, MaxPrintDepth: 5}
		if in.maxDepth != 0 {
			config.MaxDepth = in.maxDepth
		}
		if in.maxPrintDepth != 0 {
			config.MaxPrintDepth = in.maxPrintDepth
		}
		var out bytes.Buffer
		core := New(&out, config)
		core.clearTerminal = "\n------\n"
		core.maxInt = 1000
		core.minInt = -1000
		if in.maxDepth > core.maxInt {
			core.maxInt = 2 * in.maxDepth
			core.minInt = -2 * in.maxDepth
		}
		if in.board != nil {
			core.board = in.board
		}
		core.Solve()
		if err := os.WriteFile(filepath.Join("testdata", in.name+".txt"), out.Bytes(), 0644); err != nil {
			t.Errorf("Solve %v got err %v", in, err)
		}
	}
}
