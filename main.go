// Package main solves tinyhouse chess variant.
package main

import (
	"os"

	"github.com/kssilveira/chess-solver/core"
)

func main() {
	core := core.New(os.Stdout)
	core.Solve()
}
