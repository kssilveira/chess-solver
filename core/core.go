// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"slices"
	"time"
)

// Move contains a move.
type Move int16

func NewMove(fx, fy, tx, ty Move, isKing, isCapture bool) Move {
	res := Move(0)
	res |= (fx & 0b11) | ((fy & 0b11) << 2) | ((tx & 0b11) << 4) | ((ty & 0b11) << 6)
	if isKing {
		res |= 1 << 8
	}
	if isCapture {
		res |= 1 << 9
	}
	return res
}

func (m Move) Get() (Move, Move, Move, Move) {
	return m & 0b11, (m & 0b1100) >> 2, (m & 0b110000) >> 4, (m & 0b11000000) >> 6
}

func (m Move) IsKing() bool {
	return m&(1<<8) != 0
}

func (m Move) IsCapture() bool {
	return m&(1<<9) != 0
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
	board         [4][4]byte
	turn          int8
	clearTerminal string
	maxInt        int
	minInt        int
	visited       []map[[4][4]byte]int8
	solved        []map[[4][4]byte]int
	solvedMove    []map[[4][4]byte]Move
	depth         int
}

const (
	deltaDefault = iota
	deltaEmpty
	deltaEnemy
	deltaOtherEmpty
)

var (
	colors = map[byte]int8{
		byte(' '): -1,
		byte('P'): 0, byte('K'): 0, byte('R'): 0, byte('N'): 0, byte('B'): 0, byte('X'): 0,
		byte('p'): 1, byte('k'): 1, byte('r'): 1, byte('n'): 1, byte('b'): 1, byte('x'): 1,
	}
	deltas = map[byte][][]int8{
		byte('P'): [][]int8{{-1, 0, deltaEmpty}, {-1, -1, deltaEnemy}, {-1, 1, deltaEnemy}},
		byte('R'): [][]int8{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('B'): [][]int8{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('K'): [][]int8{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('N'): [][]int8{
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
		byte('p'): [][]int8{{1, 0, deltaEmpty}, {1, -1, deltaEnemy}, {1, 1, deltaEnemy}},
		byte('r'): [][]int8{{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		byte('b'): [][]int8{{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('k'): [][]int8{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		byte('n'): [][]int8{
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
	}
)

// New creates a new core.
func New(writer io.Writer, config Config) *Core {
	res := &Core{writer: writer, config: config, board: [4][4]byte{
		[4]byte([]byte("bnrk")),
		[4]byte([]byte("   p")),
		[4]byte([]byte("P   ")),
		[4]byte([]byte("KRNB")),
	},
		visited: []map[[4][4]byte]int8{{}, {}},
		solved:  []map[[4][4]byte]int{{}, {}}, solvedMove: []map[[4][4]byte]Move{{}, {}},
		clearTerminal: "\033[H\033[2J", maxInt: math.MaxInt - 1, minInt: math.MinInt + 1}
	if len(config.Board) > 1 {
		for i, row := range config.Board {
			res.board[i] = [4]byte([]byte(row))
		}
	}
	return res
}

// Solve solves the board.
func (c *Core) Solve() {
	c.depth = 0
	c.turn = 0
	c.solve()
	if c.config.MaxPrintDepth > 0 {
		c.show()
	}
}

func (c *Core) solve() int {
	c.print("after move", c.minInt, PrintConfig{ClearTerminal: true})
	if c.depth >= c.config.MaxDepth {
		res := 0
		c.print("max depth", res, PrintConfig{})
		return res
	}
	nextTurn := int8((c.turn + 1) % 2)
	moves := make([]Move, 0, 10)
	c.moves(&moves, nextTurn)
	c.sort(&moves)
	return c.move(nextTurn, moves)
}

func (c *Core) print(message string, res int, cfg PrintConfig) {
	if c.config.MaxPrintDepth != 0 && c.depth > c.config.MaxPrintDepth {
		return
	}
	fmt.Fprintf(c.writer, "\n%s\n", message)
	fmt.Fprintf(c.writer, "turn: %d\n", c.turn)
	fmt.Fprintf(c.writer, "depth: %d\n", c.depth)
	fmt.Fprintf(c.writer, "res: %d\n", res)
	if cfg.Move != 0 {
		fx, fy, tx, ty := cfg.Move.Get()
		fmt.Fprintf(c.writer, "move: %s (%d, %d) => %s (%d, %d)\n", c.what(fx, fy), fx, fy, c.what(tx, ty), tx, ty)
	}
	fmt.Fprintln(c.writer, "______")
	fmt.Fprintln(c.writer, "|"+string(bytes.Join(toBytes(c.board), []byte("|\n|")))+"|")
	fmt.Fprintln(c.writer, "‾‾‾‾‾‾")
	if cfg.ClearTerminal {
		time.Sleep(c.config.SleepDuration)
		fmt.Fprint(c.writer, c.clearTerminal)
	}
}

func (c *Core) what(x, y Move) string {
	return string(c.board[x][y])
}

func toBytes(board [4][4]byte) [][]byte {
	res := [][]byte{}
	for _, row := range board {
		one := []byte{}
		for _, v := range row {
			one = append(one, v)
		}
		res = append(res, one)
	}
	return res
}

func (c *Core) moves(moves *[]Move, nextTurn int8) {
	for i := int8(0); i < 4; i++ {
		for j := int8(0); j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			c.deltas(moves, nextTurn, i, j)
		}
	}
}

func (c *Core) deltas(moves *[]Move, nextTurn, i, j int8) {
	piece := c.board[i][j]
	for _, delta := range deltas[piece] {
		kind := int8(0)
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
		*moves = append(*moves, NewMove(Move(i), Move(j), Move(ni), Move(nj), c.board[ni][nj] == 'k' || c.board[ni][nj] == 'K', c.board[ni][nj] != ' '))
	}
}

func (c *Core) sort(moves *[]Move) {
	slices.SortFunc(*moves, func(i, j Move) int {
		if i.IsKing() {
			return -1
		}
		if j.IsKing() {
			return 1
		}
		if i.IsCapture() {
			return -1
		}
		if j.IsCapture() {
			return 1
		}
		return 0
	})
}

func (c *Core) move(nextTurn int8, moves []Move) int {
	res := c.minInt
	if len(moves) == 0 {
		res = 0
		c.print("stalemate", res, PrintConfig{})
		return res
	}
	for _, move := range moves {
		if move.IsKing() {
			res = c.maxInt - c.depth
			c.print("dead king", res, PrintConfig{Move: move})
			return res
		}
		c.print("before move", res, PrintConfig{Move: move})
		fx, fy, tx, ty := move.Get()
		what := c.board[tx][ty]
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
		c.visited[c.turn][c.board]++
		next := 0
		if nextResult, ok := c.solved[c.turn][c.board]; ok {
			next = nextResult
			c.print("solved[]", next, PrintConfig{Move: move})
		} else if c.visited[c.turn][c.board] < 3 {
			prevTurn := c.turn
			c.turn = nextTurn
			c.depth++
			next = -c.solve()
			c.depth--
			c.turn = prevTurn
			c.print("solve()", next, PrintConfig{Move: move})
			c.solved[c.turn][c.board] = next
		} else {
			c.print("repeated", next, PrintConfig{Move: move})
		}
		c.visited[c.turn][c.board]--
		c.board[fx][fy] = c.board[tx][ty]
		c.board[tx][ty] = what
		if next > res {
			res = next
			c.solvedMove[c.turn][c.board] = move
			c.print("updated res", res, PrintConfig{Move: move})
		}
	}
	c.print("final res", res, PrintConfig{Move: c.solvedMove[c.turn][c.board]})
	return res
}

func (c *Core) show() {
	res := 123456789
	c.print("show", res, PrintConfig{})
	for i := 0; i < 10; i++ {
		move := c.solvedMove[c.turn][c.board]
		if move == 0 {
			break
		}
		c.print("before move", res, PrintConfig{Move: move})

		fx, fy, tx, ty := move.Get()
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
		// c.depth++
		res = c.solved[c.turn][c.board]
		c.turn = (c.turn + 1) % 2
		c.print("after move", res, PrintConfig{Move: move})
	}
}
