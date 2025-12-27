// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/kssilveira/chess-solver/config"
	"github.com/kssilveira/chess-solver/move"
	"github.com/kssilveira/chess-solver/printconfig"
	"github.com/kssilveira/chess-solver/state"
)

// Core contains the core logic.
type Core struct {
	writer        io.Writer
	config        config.Config
	board         [4][4]byte
	turn          int8
	clearTerminal string
	visited       []map[[4][4]byte]interface{}
	solved        []map[[4][4]byte]int8
	solvedMove    []map[[4][4]byte]move.Move
	depth         int
	states        []*state.State
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
func New(writer io.Writer, config config.Config) *Core {
	res := &Core{writer: writer, config: config, board: [4][4]byte{
		[4]byte([]byte("bnrk")),
		[4]byte([]byte("   p")),
		[4]byte([]byte("P   ")),
		[4]byte([]byte("KRNB")),
	},
		visited: []map[[4][4]byte]interface{}{{}, {}},
		solved:  []map[[4][4]byte]int8{{}, {}}, solvedMove: []map[[4][4]byte]move.Move{{}, {}},
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

func (c *Core) solve() *state.State {
	if c.config.PrintDepth && c.depth%100000 == 0 {
		fmt.Fprintf(c.writer, "%d\n", c.depth)
	}
	if len(c.states) == c.depth {
		c.states = append(c.states, &state.State{})
	}
	state := c.states[c.depth]
	if c.config.EnablePrint {
		c.print("after move", -1, printconfig.PrintConfig{ClearTerminal: true})
	}
	if c.config.MaxDepth >= 0 && c.depth >= c.config.MaxDepth {
		state.Value = 0
		if c.config.EnablePrint {
			c.print("max depth", state.Value, printconfig.PrintConfig{})
		}
		return state
	}
	if cap(state.Moves) == 0 {
		state.Moves = make([]move.Move, 0, 10)
	} else {
		state.Moves = state.Moves[:0]
	}
	c.moves(&state.Moves)
	c.sort(&state.Moves)
	c.move(state)
	return state
}

func (c *Core) print(message string, res int8, cfg printconfig.PrintConfig) {
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

func (c *Core) moves(moves *[]move.Move) {
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

func (c *Core) deltas(moves *[]move.Move, i, j int8) {
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
		*moves = append(*moves, move.NewMove(move.Move(i), move.Move(j), move.Move(ni), move.Move(nj), c.board[ni][nj] == 'k' || c.board[ni][nj] == 'K', c.board[ni][nj] != ' '))
	}
}

func (c *Core) sort(moves *[]move.Move) {
	slices.SortFunc(*moves, func(i, j move.Move) int {
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

func (c *Core) move(state *state.State) {
	state.Value = -1
	if len(state.Moves) == 0 {
		state.Value = 0
		if c.config.EnablePrint {
			c.print("stalemate", state.Value, printconfig.PrintConfig{})
		}
		return
	}
	for state.MoveIndex = 0; state.MoveIndex < int8(len(state.Moves)); state.MoveIndex++ {
		if state.Moves[state.MoveIndex].IsKing() {
			state.Value = 1
			c.solvedMove[c.turn][c.board] = state.Moves[state.MoveIndex]
			if c.config.EnablePrint {
				c.print("dead king", state.Value, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			return
		}
		if c.config.EnablePrint {
			c.print("before move", state.Value, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
		}
		state.What = c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()]
		c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = ' '
		state.Next = 0
		if _, ok := c.solved[c.turn][c.board]; ok {
			state.Next = c.solved[c.turn][c.board]
			if c.config.EnablePrint {
				c.print("solved[]", state.Next, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
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
				c.print("solve()", state.Next, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			c.solved[c.turn][c.board] = state.Next
		} else {
			if c.config.EnablePrint {
				c.print("repeated", state.Next, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
		}
		c.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		c.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = state.What
		if state.Next > state.Value {
			state.Value = state.Next
			c.solvedMove[c.turn][c.board] = state.Moves[state.MoveIndex]
			if c.config.EnablePrint {
				c.print("updated res", state.Value, printconfig.PrintConfig{Move: state.Moves[state.MoveIndex]})
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
		c.print("final res", state.Value, printconfig.PrintConfig{Move: c.solvedMove[c.turn][c.board]})
	}
}

func (c *Core) show() {
	c.config.MaxPrintDepth = 0
	res := int8(123)
	c.print("show", res, printconfig.PrintConfig{})
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
		c.print("before move", res, printconfig.PrintConfig{Move: move})

		fx, fy, tx, ty := move.Get()
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
		c.depth++
		res = c.solved[c.turn][c.board]
		c.turn = (c.turn + 1) % 2
		c.print("after move", res, printconfig.PrintConfig{Move: move})
	}
}

// Play plays a game agains the solution.
func (c *Core) Play() {
	c.config.MaxPrintDepth = 0
	res := int8(123)
	c.print("play", res, printconfig.PrintConfig{})
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
		c.print("before move", res, printconfig.PrintConfig{Move: move})

		fx, fy, tx, ty := move.Get()
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
		c.depth++
		res = c.solved[c.turn][c.board]
		c.print("after move", res, printconfig.PrintConfig{Move: move})

		fmt.Printf("> ")
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
	}
}
