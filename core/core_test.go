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
		board         [4][4]byte
		maxDepth      int
		maxPrintDepth int
	}{
		{name: "default", maxDepth: 2},
		{name: "empty", board: [4][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
		}},
		{name: "P1", board: [4][4]byte{
			[4]byte([]byte("   p")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("P   ")),
		}},
		{name: "P2", board: [4][4]byte{
			[4]byte([]byte("  p ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte(" P  ")),
		}},
		{name: "P3", board: [4][4]byte{
			[4]byte([]byte(" p  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("  P ")),
		}},
		{name: "P4", board: [4][4]byte{
			[4]byte([]byte("p   ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("   P")),
		}},
		{name: "PX", board: [4][4]byte{
			[4]byte([]byte("xxx ")),
			[4]byte([]byte(" P  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
		}},
		{name: "R", board: [4][4]byte{
			[4]byte([]byte("   r")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("R   ")),
		}},
		{name: "B", board: [4][4]byte{
			[4]byte([]byte("   b")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("B   ")),
		}},
		{name: "K", board: [4][4]byte{
			[4]byte([]byte("   k")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("K   ")),
		}},
		{name: "Kk", board: [4][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("  k ")),
			[4]byte([]byte(" K  ")),
			[4]byte([]byte("    ")),
		}},
		{name: "Kk2", board: [4][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte(" k  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("K k ")),
		}},
		{name: "NB", board: [4][4]byte{
			[4]byte([]byte("nx  ")),
			[4]byte([]byte("X   ")),
			[4]byte([]byte("   x")),
			[4]byte([]byte("  XN")),
		}},
		{name: "N", board: [4][4]byte{
			[4]byte([]byte("nx  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("  XN")),
		}},
		{name: "Nk", maxDepth: 3, board: [4][4]byte{
			[4]byte([]byte("k R ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("RR  ")),
			[4]byte([]byte("   N")),
		}},
		{name: "RNk", maxDepth: 500000, maxPrintDepth: 2, board: [4][4]byte{
			[4]byte([]byte("  R ")),
			[4]byte([]byte("k   ")),
			[4]byte([]byte(" R  ")),
			[4]byte([]byte("R  N")),
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
		if in.name != "default" {
			core.board = in.board
		}
		core.Solve()
		if err := os.WriteFile(filepath.Join("testdata", in.name+".txt"), out.Bytes(), 0644); err != nil {
			t.Errorf("Solve %v got err %v", in, err)
		}
	}
}

func BenchmarkSolve(b *testing.B) {
	inputs := []struct {
		name string
	}{
		{name: "default"},
	}
	for _, in := range inputs {
		config := Config{MaxDepth: 5, MaxPrintDepth: -1}
		var out bytes.Buffer
		core := New(&out, config)
		b.Run(in.name, func(b *testing.B) {
			for b.Loop() {
				core.Solve()
			}
		})
	}
}
