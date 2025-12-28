// Package main solves tinyhouse chess variant.
package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/kssilveira/chess-solver/config"
	"github.com/kssilveira/chess-solver/core"
)

func main() {
	sleepDuration := flag.Duration("sleep_duration", 0*time.Second, "sleep duration")
	board := flag.String("board", "", "board")
	maxPrintDepth := flag.Int("max_print_depth", 1, "max depth")
	enablePlay := flag.Bool("enable_play", true, "enable play")
	printDepth := flag.Bool("print_depth", true, "print depth")
	enablePromotion := flag.Bool("enable_promotion", true, "enable promotion")
	flag.Parse()
	core := core.New(os.Stdout, config.Config{
		SleepDuration: *sleepDuration, MaxPrintDepth: *maxPrintDepth, PrintDepth: *printDepth,
		EnablePromotion: *enablePromotion,
		Board:           strings.Split(*board, ","),
	})
	core.Solve()
	if *enablePlay {
		core.Play()
	}
}
