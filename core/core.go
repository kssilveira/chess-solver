// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"
	"time"
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

// Config contains configuration.
type Config struct {
	MaxDepth      int
	SleepDuration time.Duration
	Board         []string
}

// Core contains the core logic.
type Core struct {
	writer        io.Writer
	config        Config
	board         [][]byte
	turn          int
	clearTerminal string
	maxInt        int
	minInt        int
	visited       []map[string]int
	solved        []map[string]int
	maxPrintDepth int
	depth         int
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
func New(writer io.Writer, config Config) *Core {
	res := &Core{writer: writer, config: config, turn: 1, board: [][]byte{
		[]byte("bnrk"),
		[]byte("   p"),
		[]byte("P   "),
		[]byte("KRNB"),
	}, visited: []map[string]int{{}, {}, {}}, solved: []map[string]int{{}, {}, {}}, clearTerminal: "\033[H\033[2J", maxInt: math.MaxInt, minInt: math.MinInt}
	if len(config.Board) > 1 {
		for i, row := range config.Board {
			res.board[i] = []byte(row)
		}
	}
	return res
}

// Solve solves the board.
func (c *Core) Solve() {
	c.depth = 0
	c.solve()
}

func (c *Core) solve() int {
	c.print("after move", c.minInt, Move{})
	time.Sleep(c.config.SleepDuration)
	fmt.Fprint(c.writer, c.clearTerminal)
	if c.depth >= c.config.MaxDepth {
		res := 0
		c.print("max depth", res, Move{})
		return res
	}
	nextTurn := (c.turn % 2) + 1
	moves := c.moves(nextTurn)
	moves = c.sort(moves)
	return c.move(nextTurn, moves)
}

func (c *Core) print(message string, res int, move Move) {
	if c.maxPrintDepth > 0 && c.depth > c.maxPrintDepth {
		return
	}
	fmt.Fprintf(c.writer, "\n%s\n", message)
	fmt.Fprintf(c.writer, "depth: %d\n", c.depth)
	fmt.Fprintf(c.writer, "res: %d\n", res)
	if move != (Move{}) {
		fmt.Fprintf(c.writer, "move: %s (%d, %d) => %s (%d, %d)\n", string(move.From.What), move.From.X, move.From.Y, string(move.To.What), move.To.X, move.To.Y)
	}
	fmt.Fprintln(c.writer, "______")
	fmt.Fprintln(c.writer, "|"+string(bytes.Join(c.board, []byte("|\n|")))+"|")
	fmt.Fprintln(c.writer, "‾‾‾‾‾‾")
}

func (c *Core) moves(nextTurn int) []Move {
	moves := []Move{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			moves = append(moves, c.deltas(nextTurn, i, j)...)
		}
	}
	return moves
}

func (c *Core) deltas(nextTurn, i, j int) []Move {
	moves := []Move{}
	piece := c.board[i][j]
	for _, delta := range deltas[piece] {
		kind := 0
		if len(delta) > 2 {
			kind = delta[2]
		}
		ni := i + delta[0]
		nj := j + delta[1]
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
		if kind == kindOtherEmpty && c.board[i+delta[3]][j+delta[4]] != ' ' {
			continue
		}
		moves = append(moves, Move{
			From: Point{What: piece, X: i, Y: j},
			To:   Point{What: c.board[ni][nj], X: ni, Y: nj}})
	}
	return moves
}

func (c *Core) sort(moves []Move) []Move {
	sort.Slice(moves, func(i, j int) bool {
		if moves[i].To.What == 'k' || moves[i].To.What == 'K' {
			return true
		}
		if moves[j].To.What == 'k' || moves[j].To.What == 'K' {
			return false
		}
		return i < j
	})
	return moves
}

func (c *Core) move(nextTurn int, moves []Move) int {
	res := c.minInt
	resMove := Move{}
	if len(moves) == 0 {
		res = 0
		c.print("stalemate", res, Move{})
		return res
	}
	for _, move := range moves {
		if move.To.What == 'k' || move.To.What == 'K' {
			res = c.maxInt - c.depth
			c.print("dead king", res, move)
			return res
		}
		c.print("before move", res, move)
		c.board[move.To.X][move.To.Y] = c.board[move.From.X][move.From.Y]
		c.board[move.From.X][move.From.Y] = ' '
		key := string(bytes.Join(c.board, nil))
		c.visited[c.turn][key]++
		next := 0
		if _, ok := c.solved[c.turn][key]; ok {
			next = c.solved[c.turn][key]
		} else if c.visited[c.turn][key] < 3 {
			prevTurn := c.turn
			c.turn = nextTurn
			c.depth++
			next = -c.solve()
			c.depth--
			c.turn = prevTurn
			c.solved[c.turn][key] = next
		}
		c.visited[c.turn][key]--
		c.board[move.To.X][move.To.Y] = move.To.What
		c.board[move.From.X][move.From.Y] = move.From.What
		if next > res {
			res = next
			resMove = move
			c.print("updated res", res, resMove)
		}
	}
	c.print("final res", res, resMove)
	return res
}
