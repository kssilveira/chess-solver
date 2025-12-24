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
	board    [][]byte
	writer   io.Writer
	turn     int
	maxDepth int
}

var (
	colors = map[byte]int{
		byte(' '): 0,
		byte('P'): 1, byte('K'): 1, byte('R'): 1, byte('N'): 1, byte('B'): 1,
		byte('p'): 2, byte('k'): 2, byte('r'): 2, byte('n'): 2, byte('b'): 2,
	}
	deltas = map[byte][][]int{
		byte('P'): [][]int{{-1, 0, 1}, {-1, -1, 2}, {-1, 1, 2}},
		byte('R'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('B'): [][]int{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
	}
)

// New creates a new core.
func New(writer io.Writer) *Core {
	return &Core{writer: writer, turn: 1, maxDepth: 3, board: [][]byte{
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
	fmt.Fprintln(c.writer, "######")
	fmt.Fprintln(c.writer, "#"+string(bytes.Join(c.board, []byte("#\n#")))+"#")
	fmt.Fprintln(c.writer, "######")
	if depth >= c.maxDepth {
		return
	}
	moves := []Move{}
	nextTurn := (c.turn % 2) + 1
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			for _, delta := range deltas[piece] {
				dx := delta[0]
				dy := delta[1]
				kind := 0
				if len(delta) > 2 {
					kind = delta[2]
				}
				ni := i + dx
				nj := j + dy
				if ni < 0 || ni >= 4 || nj < 0 || nj >= 4 {
					continue
				}
				if kind == 0 && c.board[ni][nj] != ' ' && colors[c.board[ni][nj]] != nextTurn {
					continue
				}
				if kind == 1 && c.board[ni][nj] != ' ' {
					continue
				}
				if kind == 2 && colors[c.board[ni][nj]] != nextTurn {
					continue
				}
				moves = append(moves, Move{
					From: Point{What: piece, X: i, Y: j},
					To:   Point{What: c.board[ni][nj], X: ni, Y: nj}})
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
