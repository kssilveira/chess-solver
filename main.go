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
	maxDepth := flag.Int("max_depth", -1, "max depth")
	sleepDuration := flag.Duration("sleep_duration", 0*time.Second, "sleep duration")
	board := flag.String("board", "", "board")
	maxPrintDepth := flag.Int("max_print_depth", 1, "max depth")
	enablePlay := flag.Bool("enable_play", true, "enable play")
	flag.Parse()
	core := core.New(os.Stdout, core.Config{
		MaxDepth: *maxDepth, SleepDuration: *sleepDuration, MaxPrintDepth: *maxPrintDepth,
		Board:       strings.Split(*board, ","),
		EnablePrint: true,
	})
	core.Solve()
	if *enablePlay {
		core.Play()
	}
}
