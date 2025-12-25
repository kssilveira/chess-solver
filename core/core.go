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
	MaxPrintDepth int
}

// PrintConfig contains print configuration.
type PrintConfig struct {
	Move          Move
	ClearTerminal bool
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
	solvedMove    []map[string]Move
	depth         int
}

const (
	deltaDefault = iota
	deltaEmpty
	deltaEnemy
	deltaOtherEmpty
)

var (
	colors = map[byte]int{
		byte(' '): -1,
		byte('P'): 0, byte('K'): 0, byte('R'): 0, byte('N'): 0, byte('B'): 0, byte('X'): 0,
		byte('p'): 1, byte('k'): 1, byte('r'): 1, byte('n'): 1, byte('b'): 1, byte('x'): 1,
	}
	deltas = map[byte][][]int{
		byte('P'): [][]int{{-1, 0, deltaEmpty}, {-1, -1, deltaEnemy}, {-1, 1, deltaEnemy}},
		byte('R'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('B'): [][]int{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('K'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('N'): [][]int{
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
		byte('p'): [][]int{{1, 0, deltaEmpty}, {1, -1, deltaEnemy}, {1, 1, deltaEnemy}},
		byte('r'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('b'): [][]int{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('k'): [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('n'): [][]int{
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
	}
)

// New creates a new core.
func New(writer io.Writer, config Config) *Core {
	res := &Core{writer: writer, config: config, board: [][]byte{
		[]byte("bnrk"),
		[]byte("   p"),
		[]byte("P   "),
		[]byte("KRNB"),
	},
		visited: []map[string]int{{}, {}},
		solved:  []map[string]int{{}, {}}, solvedMove: []map[string]Move{{}, {}},
		clearTerminal: "\033[H\033[2J", maxInt: math.MaxInt - 1, minInt: math.MinInt + 1}
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
	c.turn = 0
	c.solve()
	c.show()
}

func (c *Core) solve() int {
	c.print("after move", c.minInt, PrintConfig{ClearTerminal: true})
	if c.depth >= c.config.MaxDepth {
		res := 0
		c.print("max depth", res, PrintConfig{})
		return res
	}
	nextTurn := (c.turn + 1) % 2
	moves := c.moves(nextTurn)
	moves = c.sort(moves)
	return c.move(nextTurn, moves)
}

func (c *Core) print(message string, res int, cfg PrintConfig) {
	if c.config.MaxPrintDepth > 0 && c.depth > c.config.MaxPrintDepth {
		return
	}
	fmt.Fprintf(c.writer, "\n%s\n", message)
	fmt.Fprintf(c.writer, "turn: %d\n", c.turn)
	fmt.Fprintf(c.writer, "depth: %d\n", c.depth)
	fmt.Fprintf(c.writer, "res: %d\n", res)
	if cfg.Move != (Move{}) {
		fmt.Fprintf(c.writer, "move: %s (%d, %d) => %s (%d, %d)\n", string(cfg.Move.From.What), cfg.Move.From.X, cfg.Move.From.Y, string(cfg.Move.To.What), cfg.Move.To.X, cfg.Move.To.Y)
	}
	fmt.Fprintln(c.writer, "______")
	fmt.Fprintln(c.writer, "|"+string(bytes.Join(c.board, []byte("|\n|")))+"|")
	fmt.Fprintln(c.writer, "‾‾‾‾‾‾")
	if cfg.ClearTerminal {
		time.Sleep(c.config.SleepDuration)
		fmt.Fprint(c.writer, c.clearTerminal)
	}
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
		if (kind == deltaDefault || kind == deltaOtherEmpty) && c.board[ni][nj] != ' ' && colors[c.board[ni][nj]] != nextTurn {
			continue
		}
		if kind == deltaEmpty && c.board[ni][nj] != ' ' {
			continue
		}
		if kind == deltaEnemy && colors[c.board[ni][nj]] != nextTurn {
			continue
		}
		if kind == deltaOtherEmpty && c.board[i+delta[3]][j+delta[4]] != ' ' {
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
		if moves[i].To.What != ' ' {
			return true
		}
		if moves[j].To.What != ' ' {
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
		c.print("stalemate", res, PrintConfig{})
		return res
	}
	for _, move := range moves {
		if move.To.What == 'k' || move.To.What == 'K' {
			res = c.maxInt - c.depth
			c.print("dead king", res, PrintConfig{Move: move})
			return res
		}
		c.print("before move", res, PrintConfig{Move: move})
		c.board[move.To.X][move.To.Y] = c.board[move.From.X][move.From.Y]
		c.board[move.From.X][move.From.Y] = ' '
		key := string(bytes.Join(c.board, nil))
		c.visited[c.turn][key]++
		next := 0
		if nextResult, ok := c.solved[c.turn][key]; ok {
			next = nextResult
			c.print("solved[]", next, PrintConfig{Move: move})
		} else if c.visited[c.turn][key] < 3 {
			prevTurn := c.turn
			c.turn = nextTurn
			c.depth++
			next = -c.solve()
			c.depth--
			c.turn = prevTurn
			c.print("solve()", next, PrintConfig{Move: move})
			c.solved[c.turn][key] = next
		} else {
			c.print("repeated", next, PrintConfig{Move: move})
		}
		c.visited[c.turn][key]--
		c.board[move.To.X][move.To.Y] = move.To.What
		c.board[move.From.X][move.From.Y] = move.From.What
		if next > res {
			res = next
			resMove = move
			c.print("updated res", res, PrintConfig{Move: resMove})
		}
	}
	c.print("final res", res, PrintConfig{Move: resMove})
	key := string(bytes.Join(c.board, nil))
	c.solvedMove[c.turn][key] = resMove
	return res
}

func (c *Core) show() {
	res := 123456789
	c.print("show", res, PrintConfig{})
	key := string(bytes.Join(c.board, nil))
	for i := 0; i < 10; i++ {
		move := c.solvedMove[c.turn][key]
		if move == (Move{}) {
			break
		}
		c.print("before move", res, PrintConfig{Move: move})

		c.board[move.To.X][move.To.Y] = c.board[move.From.X][move.From.Y]
		c.board[move.From.X][move.From.Y] = ' '
		// c.depth++
		key = string(bytes.Join(c.board, nil))
		res = c.solved[c.turn][key]
		c.turn = (c.turn + 1) % 2
		c.print("after move", res, PrintConfig{Move: move, ClearTerminal: true})
	}
}
