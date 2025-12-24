// Package main solves tinyhouse chess variant.
package main

import (
	"flag"
	"os"
	"time"

	"github.com/kssilveira/chess-solver/core"
)

func main() {
	maxDepth := flag.Int("max_depth", 2, "max depth")
	sleepDuration := flag.Duration("sleep_duration", time.Second, "sleep duration")
	flag.Parse()
	core := core.New(os.Stdout, core.Config{MaxDepth: *maxDepth, SleepDuration: *sleepDuration})
	core.Solve()
}
