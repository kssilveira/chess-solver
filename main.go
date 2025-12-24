// Package main solves tinyhouse chess variant.
package main

import (
	"os"
	"time"

	"github.com/kssilveira/chess-solver/core"
)

func main() {
	core := core.New(os.Stdout)
	core.MaxDepth = 2
	core.SleepDuration = 1 * time.Second
	core.Solve()
}
