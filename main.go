// Package main solves tinyhouse chess variant.
package main

import (
	"fmt"
	"strings"
)

type Point struct {
	X int
	Y int
}

type Move struct {
	What string
	From Point
	To   Point
}

func main() {
	board := []string{
		"bnrk",
		"   p",
		"P   ",
		"KRNB",
	}
	fmt.Println(strings.Join(board, "\n"))
	moves := []Move{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if board[i][j] == 'P' {
				ni := i - 1
				nj := j
				if board[ni][nj] == ' ' {
					moves = append(moves, Move{What: string(board[i][j]), From: Point{X: i, Y: j}, To: Point{X: ni, Y: nj}})
				}
			}
		}
	}
	for _, move := range moves {
		fmt.Println(move)
	}
}
