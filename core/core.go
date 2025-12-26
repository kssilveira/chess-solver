// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"time"
)

// Move contains a move.
type Move int16

// NewMove creates a new move.
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

// Get gets coordinates.
func (m Move) Get() (int8, int8, int8, int8) {
	return m.FromX(), m.FromY(), m.ToX(), m.ToY()
}

// FromX returns from x.
func (m Move) FromX() int8 {
	return int8(m & 0b11)
}

// FromY returns from y.
func (m Move) FromY() int8 {
	return int8((m & 0b1100) >> 2)
}

// ToX returns to x.
func (m Move) ToX() int8 {
	return int8((m & 0b110000) >> 4)
}

// ToY returns to y.
func (m Move) ToY() int8 {
	return int8((m & 0b11000000) >> 6)
}

// IsKing returns is king.
func (m Move) IsKing() bool {
	return m&(1<<8) != 0
}

// IsCapture returns is capture.
func (m Move) IsCapture() bool {
	return m&(1<<9) != 0
}

// Config contains configuration.
type Config struct {
	MaxDepth      int
	SleepDuration time.Duration
	Board         []string
	MaxPrintDepth int
	EnablePrint   bool
	EnableShow    bool
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
	visited       []map[[4][4]byte]interface{}
	solved        []map[[4][4]byte]int8
	solvedMove    []map[[4][4]byte]Move
	depth         int
}

// State contains the recursion state.
type State struct {
	Value     int8
	Next      int8
	MoveIndex int8
	What      byte
	Moves     []Move
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
		visited: []map[[4][4]byte]interface{}{{}, {}},
		solved:  []map[[4][4]byte]int8{{}, {}}, solvedMove: []map[[4][4]byte]Move{{}, {}},
		clearTerminal: "\033[H\033[2J"}
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
	if c.config.EnablePrint && c.config.EnableShow {
		c.show()
	}
}

func (c *Core) solve() *State {
	state := &State{}
	if c.config.EnablePrint {
		c.print("after move", -1, PrintConfig{ClearTerminal: true})
	}
	if c.config.MaxDepth >= 0 && c.depth >= c.config.MaxDepth {
		state.Value = 0
		if c.config.EnablePrint {
			c.print("max depth", state.Value, PrintConfig{})
		}
		return state
	}
	state.Moves = make([]Move, 0, 10)
	c.moves(&state.Moves)
	c.sort(&state.Moves)
	c.move(state)
	return state
}

func (c *Core) print(message string, res int8, cfg PrintConfig) {
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

func (c *Core) what(x, y int8) string {
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

func (c *Core) moves(moves *[]Move) {
	for i := int8(0); i < 4; i++ {
		for j := int8(0); j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			c.deltas(moves, i, j)
		}
	}
}

func (c *Core) deltas(moves *[]Move, i, j int8) {
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
		nextTurn := int8((c.turn + 1) % 2)
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

func (c *Core) move(state *State) {
	state.Value = -1
	if len(state.Moves) == 0 {
		state.Value = 0
		if c.config.EnablePrint {
			c.print("stalemate", state.Value, PrintConfig{})
		}
		return
	}
	for state.MoveIndex = 0; state.MoveIndex < int8(len(state.Moves)); state.MoveIndex++ {
		if state.Moves[state.MoveIndex].IsKing() {
			state.Value = 1
			c.solvedMove[c.turn][c.board] = state.Moves[state.MoveIndex]
			if c.config.EnablePrint {
				c.print("dead king", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			return
		}
		if c.config.EnablePrint {
			c.print("before move", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
		}
		state.What = c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()]
		c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = ' '
		state.Next = 0
		if _, ok := c.solved[c.turn][c.board]; ok {
			state.Next = c.solved[c.turn][c.board]
			if c.config.EnablePrint {
				c.print("solved[]", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
		} else if _, ok := c.visited[c.turn][c.board]; !ok {
			c.visited[c.turn][c.board] = struct{}{}
			c.turn = int8((c.turn + 1) % 2)
			c.depth++
			nextState := c.solve()
			state.Next = -nextState.Value
			c.depth--
			c.turn = int8((c.turn + 1) % 2)
			if c.config.EnablePrint {
				c.print("solve()", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			c.solved[c.turn][c.board] = state.Next
		} else {
			if c.config.EnablePrint {
				c.print("repeated", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
		}
		c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = state.What
		if state.Next > state.Value {
			state.Value = state.Next
			c.solvedMove[c.turn][c.board] = state.Moves[state.MoveIndex]
			if c.config.EnablePrint {
				c.print("updated res", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			if state.Value == 1 {
				break
			}
		}
	}
	if state.Value == -1 {
		c.solvedMove[c.turn][c.board] = state.Moves[0]
	}
	if c.config.EnablePrint {
		c.print("final res", state.Value, PrintConfig{Move: c.solvedMove[c.turn][c.board]})
	}
}

func (c *Core) show() {
	res := int8(123)
	c.print("show", res, PrintConfig{})
	visited := []map[[4][4]byte]interface{}{{}, {}}
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
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

// Play plays a game agains the solution.
func (c *Core) Play() {
	res := int8(123)
	c.print("play", res, PrintConfig{})
	c.turn = 0
	visited := []map[[4][4]byte]interface{}{{}, {}}
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
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
		c.print("after move", res, PrintConfig{Move: move})

		fmt.Printf("> ")
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
	}
}
