// Package main solves tinyhouse chess variant.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kssilveira/chess-solver/config"
	"github.com/kssilveira/chess-solver/core"
)

func doRunAll() {
	configs := []config.Config{
		{Board: "   k,    ,P   ,KR  ", EnablePromotion: true},
		{Board: "   k,    ,P   ,KR  ", EnableDrop: true},
		{Board: "   k,    ,P   ,KR  ", EnablePromotion: true, EnableDrop: true},
		{Board: "   k,    ,P   ,KRNB"},
		// {Board: "   k,    ,P   ,KRNB", EnablePromotion: true},
		{Board: "   k,   p,P   ,KRNB"},
		// {Board: "   k,   p,P   ,KRNB", EnablePromotion: true},
		// {Board: "b  k,   p,P   ,KRNB", EnablePromotion: true},
	}
	buffers := []*bytes.Buffer{}
	var wg sync.WaitGroup
	for _, cfg := range configs {
		var buffer bytes.Buffer
		buffers = append(buffers, &buffer)
		wg.Go(func(config config.Config, buffer *bytes.Buffer) func() {
			return func() {
				config.MaxPrintDepth = -1
				core := core.New(buffer, config)
				// core := core.New(os.Stdout, config)
				core.Solve()
			}
		}(cfg, &buffer))
	}
	wg.Wait()
	for i, config := range configs {
		desc := []string{
			fmt.Sprintf("--board='%s'", config.Board),
		}
		if config.EnablePromotion {
			desc = append(desc, "--enable_promotion")
		}
		if config.EnableDrop {
			desc = append(desc, "--enable_drop")
		}
		fmt.Printf("\n%s\n%s", strings.Join(desc, " "), buffers[i].String())
	}
}

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
		doRunAll()
		return
	}
	core := core.New(os.Stdout, cfg)
	core.Solve()
	if *enablePlay {
		core.Play()
	}
}
