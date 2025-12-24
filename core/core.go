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

const (
	kindDefault = iota
	kindEmpty
	kindEnemy
	kindOtherEmpty
)

var (
	colors = map[byte]int{
		byte(' '): 0,
		byte('P'): 1, byte('K'): 1, byte('R'): 1, byte('N'): 1, byte('B'): 1, byte('X'): 1,
		byte('p'): 2, byte('k'): 2, byte('r'): 2, byte('n'): 2, byte('b'): 2, byte('x'): 2,
	}
	deltas = map[byte][][]int{
		byte('P'): [][]int{{-1, 0, kindEmpty}, {-1, -1, kindEnemy}, {-1, 1, kindEnemy}},
		byte('R'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('B'): [][]int{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('K'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('N'): [][]int{
			{-2, -1, kindOtherEmpty, -1, 0}, {-2, 1, kindOtherEmpty, -1, 0},
			{-1, -2, kindOtherEmpty, 0, -1}, {1, -2, kindOtherEmpty, 0, -1},
			{2, -1, kindOtherEmpty, 1, 0}, {2, 1, kindOtherEmpty, 1, 0},
			{-1, 2, kindOtherEmpty, 0, 1}, {1, 2, kindOtherEmpty, 0, 1},
		},
		byte('p'): [][]int{{1, 0, kindEmpty}, {1, -1, kindEnemy}, {1, 1, kindEnemy}},
		byte('r'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('b'): [][]int{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('k'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('n'): [][]int{
			{-2, -1, kindOtherEmpty, -1, 0}, {-2, 1, kindOtherEmpty, -1, 0},
			{-1, -2, kindOtherEmpty, 0, -1}, {1, -2, kindOtherEmpty, 0, -1},
			{2, -1, kindOtherEmpty, 1, 0}, {2, 1, kindOtherEmpty, 1, 0},
			{-1, 2, kindOtherEmpty, 0, 1}, {1, 2, kindOtherEmpty, 0, 1},
		},
	}
)

// New creates a new core.
func New(writer io.Writer) *Core {
	return &Core{writer: writer, turn: 1, maxDepth: 5, board: [][]byte{
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
				if (kind == kindDefault || kind == kindOtherEmpty) && c.board[ni][nj] != ' ' && colors[c.board[ni][nj]] != nextTurn {
					continue
				}
				if kind == kindEmpty && c.board[ni][nj] != ' ' {
					continue
				}
				if kind == kindEnemy && colors[c.board[ni][nj]] != nextTurn {
					continue
				}
				if kind == kindOtherEmpty {
					odx := delta[3]
					ody := delta[4]
					oni := i + odx
					onj := i + ody
					if oni < 0 || oni >= 4 || onj < 0 || onj >= 4 {
						continue
					}
					if c.board[oni][onj] != ' ' {
						continue
					}
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
		prevTurn := c.turn
		c.turn = nextTurn
		c.solve(depth + 1)
		c.turn = prevTurn
		c.board[move.To.X][move.To.Y] = move.To.What
		c.board[move.From.X][move.From.Y] = move.From.What
	}
}
