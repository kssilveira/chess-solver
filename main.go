// Package main solves tinyhouse chess variant.
package main

import (
	"bytes"
	"fmt"
)

type Point struct {
	X    int
	Y    int
	What byte
}

type Move struct {
	From Point
	To   Point
}

func solve(board [][]byte, depth int) {
	fmt.Printf("\ndepth: %d\n", depth)
	fmt.Println(string(bytes.Join(board, []byte("\n"))))
	moves := []Move{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if board[i][j] == 'P' {
				ni := i - 1
				nj := j
				if board[ni][nj] == ' ' {
					moves = append(moves, Move{
						From: Point{What: board[i][j], X: i, Y: j},
						To:   Point{What: board[ni][nj], X: ni, Y: nj}})
				}
			}
		}
	}
	for _, move := range moves {
		board[move.To.X][move.To.Y] = board[move.From.X][move.From.Y]
		board[move.From.X][move.From.Y] = ' '
		solve(board, depth+1)
		board[move.To.X][move.To.Y] = move.To.What
		board[move.From.X][move.From.Y] = move.From.What
	}
}

func main() {
	board := [][]byte{
		[]byte("bnrk"),
		[]byte("   p"),
		[]byte("P   "),
		[]byte("KRNB"),
	}
	solve(board, 0 /* depth */)
}
