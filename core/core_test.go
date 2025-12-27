package core

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSolve(t *testing.T) {
	inputs := []struct {
		name          string
		board         string
		maxDepth      int
		maxPrintDepth int
	}{
		{name: "default", maxDepth: 2},
		{name: "empty", board: "" +
			"    ," +
			"    ," +
			"    ," +
			"    ",
		},
		{name: "P1", board: "" +
			"   p," +
			"    ," +
			"    ," +
			"P   ",
		},
		{name: "P2", board: "" +
			"  p ," +
			"    ," +
			"    ," +
			" P  ",
		},
		{name: "P3", board: "" +
			" p  ," +
			"    ," +
			"    ," +
			"  P ",
		},
		{name: "P4", board: "" +
			"p   ," +
			"    ," +
			"    ," +
			"   P",
		},
		{name: "PX", board: "" +
			"xxx ," +
			" P  ," +
			"    ," +
			"    ",
		},
		{name: "R", board: "" +
			"   r," +
			"    ," +
			"    ," +
			"R   ",
		},
		{name: "B", board: "" +
			"   b," +
			"    ," +
			"    ," +
			"B   ",
		},
		{name: "K", board: "" +
			"   k," +
			"    ," +
			"    ," +
			"K   ",
		},
		{name: "Kk", board: "" +
			"    ," +
			"  k ," +
			" K  ," +
			"    ",
		},
		{name: "Kk2", board: "" +
			"    ," +
			" k  ," +
			"    ," +
			"K k ",
		},
		{name: "NB", board: "" +
			"nx  ," +
			"X   ," +
			"   x," +
			"  XN",
		},
		{name: "N", board: "" +
			"nx  ," +
			"    ," +
			"    ," +
			"  XN",
		},
		{name: "Nk", maxDepth: 3, board: "" +
			"k R ," +
			"    ," +
			"RR  ," +
			"   N",
		},
		{name: "RNk", maxDepth: 500000, maxPrintDepth: 2, board: "" +
			"  R ," +
			"k   ," +
			" R  ," +
			"R  N",
		},
	}
	for _, in := range inputs {
		config := Config{MaxDepth: 5, MaxPrintDepth: 5, EnablePrint: true, EnableShow: true, NumSolvers: 1}
		if in.name != "default" {
			config.Board = strings.Split(in.board, ",")
		}
		if in.maxDepth != 0 {
			config.MaxDepth = in.maxDepth
		}
		if in.maxPrintDepth != 0 {
			config.MaxPrintDepth = in.maxPrintDepth
		}
		var out bytes.Buffer
		core := New(&out, config)
		core.clearTerminal = "\n------\n"
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
		config := Config{MaxDepth: 5, MaxPrintDepth: -1, NumSolvers: int8(runtime.GOMAXPROCS(0))}
		var out bytes.Buffer
		core := New(&out, config)
		b.Run(in.name, func(b *testing.B) {
			for b.Loop() {
				core.Solve()
			}
		})
	}
}
