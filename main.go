// Package main solves tinyhouse chess variant.
package main

import (
	"fmt"
	"strings"
)

func main() {
	board := []string{
		"bnrk",
		"   p",
		"P   ",
		"KRNB",
	}
	fmt.Println(strings.Join(board, "\n"))
}
