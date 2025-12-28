package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kssilveira/chess-solver/config"
)

func TestSolve(t *testing.T) {
	inputs := []struct {
		name          string
		board         [6][4]byte
		maxPrintDepth int
		enableDrop    bool
	}{
		{name: "empty", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "P1", board: [6][4]byte{
			[4]byte([]byte("   p")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("P   ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "P2", board: [6][4]byte{
			[4]byte([]byte("  p ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte(" P  ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "P3", board: [6][4]byte{
			[4]byte([]byte(" p  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("  P ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "P4", board: [6][4]byte{
			[4]byte([]byte("p   ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("   P")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "PX", board: [6][4]byte{
			[4]byte([]byte("xxx ")),
			[4]byte([]byte(" P  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "R", board: [6][4]byte{
			[4]byte([]byte("   r")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("R   ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "B", board: [6][4]byte{
			[4]byte([]byte("   b")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("B   ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "K", board: [6][4]byte{
			[4]byte([]byte("   k")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("K   ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "Kk", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("  k ")),
			[4]byte([]byte(" K  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "Kk2", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte(" k  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("K k ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "NB", board: [6][4]byte{
			[4]byte([]byte("nx  ")),
			[4]byte([]byte("X   ")),
			[4]byte([]byte("   x")),
			[4]byte([]byte("  XN")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "N", board: [6][4]byte{
			[4]byte([]byte("nx  ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("  XN")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "Nk", board: [6][4]byte{
			[4]byte([]byte("k R ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("RR  ")),
			[4]byte([]byte("   N")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "RNk", maxPrintDepth: 2, board: [6][4]byte{
			[4]byte([]byte("  R ")),
			[4]byte([]byte("k   ")),
			[4]byte([]byte(" R  ")),
			[4]byte([]byte("R  N")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "PkR", board: [6][4]byte{
			[4]byte([]byte("k   ")),
			[4]byte([]byte("xxP ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "PkN", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte(" xP ")),
			[4]byte([]byte("kx  ")),
			[4]byte([]byte("xx  ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "PkB", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("x P ")),
			[4]byte([]byte("kx  ")),
			[4]byte([]byte("xx  ")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
		{name: "D1", board: [6][4]byte{
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("1111")),
			[4]byte([]byte("1000")),
		}},
	}
	for _, in := range inputs {
		config := config.Config{
			MaxPrintDepth: 5, EnableShow: true, EnablePromotion: true, EnableDrop: in.enableDrop}
		if in.maxPrintDepth != 0 {
			config.MaxPrintDepth = in.maxPrintDepth
		}
		var out bytes.Buffer
		core := New(&out, config)
		core.clearTerminal = "\n------\n"
		core.board = in.board
		core.Solve()
		if err := os.WriteFile(filepath.Join("testdata", in.name+".txt"), out.Bytes(), 0644); err != nil {
			t.Errorf("Solve %v got err %v", in, err)
		}
	}
}

func BenchmarkSolve(b *testing.B) {
	inputs := []struct {
		name  string
		board [6][4]byte
	}{
		{name: "Nk", board: [6][4]byte{
			[4]byte([]byte("k R ")),
			[4]byte([]byte("    ")),
			[4]byte([]byte("RR  ")),
			[4]byte([]byte("   N")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		}},
	}
	for _, in := range inputs {
		config := config.Config{MaxPrintDepth: -1}
		b.Run(in.name, func(b *testing.B) {
			for b.Loop() {
				var out bytes.Buffer
				core := New(&out, config)
				core.board = in.board
				core.Solve()
			}
		})
	}
}
