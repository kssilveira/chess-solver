// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
)

// Point contains a point.
type Point struct {
	X    int
	Y    int
	What byte
}

// Move contains a move.
type Move struct {
	From Point
	To   Point
}

// Core contains the core logic.
type Core struct {
	board  [][]byte
	writer io.Writer
}

// New creates a new core.
func New(writer io.Writer) *Core {
	return &Core{writer: writer, board: [][]byte{
		[]byte("bnrk"),
		[]byte("   p"),
		[]byte("P   "),
		[]byte("KRNB"),
	}}
}

// Solve solves the board.
func (c *Core) Solve() {
	c.solve(0 /* depth */)
}

func (c *Core) solve(depth int) {
	fmt.Fprintf(c.writer, "\ndepth: %d\n", depth)
	fmt.Fprintln(c.writer, string(bytes.Join(c.board, []byte("\n"))))
	moves := []Move{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if c.board[i][j] == 'P' {
				ni := i - 1
				nj := j
				if c.board[ni][nj] == ' ' {
					moves = append(moves, Move{
						From: Point{What: c.board[i][j], X: i, Y: j},
						To:   Point{What: c.board[ni][nj], X: ni, Y: nj}})
				}
			}
		}
	}
	for _, move := range moves {
		c.board[move.To.X][move.To.Y] = c.board[move.From.X][move.From.Y]
		c.board[move.From.X][move.From.Y] = ' '
		c.solve(depth + 1)
		c.board[move.To.X][move.To.Y] = move.To.What
		c.board[move.From.X][move.From.Y] = move.From.What
	}
}
