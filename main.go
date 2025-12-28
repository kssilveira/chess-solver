// Package main solves tinyhouse chess variant.
package main

import (
	"flag"
	"os"
	"time"

	"github.com/kssilveira/chess-solver/config"
	"github.com/kssilveira/chess-solver/core"
)

func main() {
	sleepDuration := flag.Duration("sleep_duration", 0*time.Second, "sleep duration")
	board := flag.String("board", "", "board")
	maxPrintDepth := flag.Int("max_print_depth", -1, "max depth")
	enablePlay := flag.Bool("enable_play", false, "enable play")
	printDepth := flag.Bool("print_depth", true, "print depth")
	enablePromotion := flag.Bool("enable_promotion", false, "enable promotion")
	enableDrop := flag.Bool("enable_drop", false, "enable drop")
	runAll := flag.Bool("run_all", false, "run all")
	flag.Parse()
	cfg := config.Config{
		SleepDuration: *sleepDuration, MaxPrintDepth: *maxPrintDepth, PrintDepth: *printDepth,
		EnablePromotion: *enablePromotion, EnableDrop: *enableDrop,
		Board: *board,
	}
	if *runAll {
		core.RunAll(os.Stdout, []config.Config{
			{Board: "   k,    ,P   ,KR  "},
			{Board: "   k,    ,P   ,KR  ", EnablePromotion: true},
			{Board: "   k,    ,P   ,KR  ", EnableDrop: true},
			//x {Board: "   k,    ,P   ,KR  ", EnablePromotion: true, EnableDrop: true},
			//x {Board: "   k,    ,P   ,KRNB"},
			// {Board: "   k,    ,P   ,KRNB", EnablePromotion: true},
			//x {Board: "   k,   p,P   ,KRNB"},
			// {Board: "   k,   p,P   ,KRNB", EnablePromotion: true},
			// {Board: "b  k,   p,P   ,KRNB"},
		})
		return
	}
	core := core.New(os.Stdout, cfg)
	core.Solve()
	if *enablePlay {
		core.Play()
	}
}
