// Package main solves tinyhouse chess variant.
package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/kssilveira/chess-solver/core"
)

func main() {
	maxDepth := flag.Int("max_depth", 2, "max depth")
	sleepDuration := flag.Duration("sleep_duration", time.Second, "sleep duration")
	board := flag.String("board", "", "board")
	flag.Parse()
	core := core.New(os.Stdout, core.Config{MaxDepth: *maxDepth, SleepDuration: *sleepDuration, Board: strings.Split(*board, ",")})
	core.Solve()
}
