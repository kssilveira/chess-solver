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
		{name: "P1", board: [][]byte{
			[]byte("    "),
			[]byte("    "),
			[]byte("    "),
			[]byte("P   "),
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
