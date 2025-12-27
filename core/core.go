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

func (c *Core) solve() {
	if c.config.PrintDepth && c.depth%100000 == 0 {
		fmt.Fprintf(c.writer, "%d\n", c.depth)
	}
	if len(c.states) == c.depth {
		c.states = append(c.states, &state.State{})
	}
	if c.config.EnablePrint {
		c.print("after move", -1, printconfig.PrintConfig{ClearTerminal: true})
	}
	if c.config.MaxDepth >= 0 && c.depth >= c.config.MaxDepth {
		c.states[c.depth].Value = 0
		if c.config.EnablePrint {
			c.print("max depth", c.states[c.depth].Value, printconfig.PrintConfig{})
		}
		return
	}
	if cap(c.states[c.depth].Moves) == 0 {
		c.states[c.depth].Moves = make([]move.Move, 0, 10)
	} else {
		c.states[c.depth].Moves = c.states[c.depth].Moves[:0]
	}
	c.moves()
	c.sort()
	c.move()
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
		fmt.Fprintf(
			c.writer,
			"move: %s (%d, %d) => %s (%d, %d)\n",
			c.what(fx, fy), fx, fy, c.what(tx, ty), tx, ty)
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

func (c *Core) moves() {
	for i := int8(0); i < 4; i++ {
		for j := int8(0); j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			c.deltas(i, j)
		}
	}
}

func (c *Core) deltas(i, j int8) {
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
		if (kind == deltaDefault || kind == deltaOtherEmpty) &&
			c.board[ni][nj] != ' ' && colors[c.board[ni][nj]] != nextTurn {
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
		c.states[c.depth].Moves = append(
			c.states[c.depth].Moves,
			move.NewMove(
				move.Move(i), move.Move(j), move.Move(ni), move.Move(nj),
				c.board[ni][nj] == 'k' || c.board[ni][nj] == 'K',
				c.board[ni][nj] != ' '))
	}
}

func (c *Core) sort() {
	slices.SortFunc(c.states[c.depth].Moves, func(i, j move.Move) int {
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

func (c *Core) move() {
	c.states[c.depth].Value = -1
	if c.staleMate() {
		return
	}
	for c.states[c.depth].MoveIndex = 0; c.states[c.depth].MoveIndex < int8(len(c.states[c.depth].Moves)); c.states[c.depth].MoveIndex++ {
		if c.deadKing() {
			return
		}
		c.doMove()
		c.states[c.depth].Next = 0
		if c.states[c.depth].Next, c.states[c.depth].OK = c.solved[c.turn][c.board]; c.states[c.depth].OK {
			if c.config.EnablePrint {
				c.print("solved[]", c.states[c.depth].Next, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
			}
		} else if _, c.states[c.depth].OK = c.visited[c.turn][c.board]; !c.states[c.depth].OK {
			c.visited[c.turn][c.board] = struct{}{}
			c.turn = int8((c.turn + 1) % 2)
			c.depth++
			c.solve()
			c.depth--
			c.states[c.depth].Next = -c.states[c.depth+1].Value
			c.turn = int8((c.turn + 1) % 2)
			c.solved[c.turn][c.board] = c.states[c.depth].Next
			if c.config.EnablePrint {
				c.print("solve()", c.states[c.depth].Next, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
			}
		} else {
			if c.config.EnablePrint {
				c.print("repeated", c.states[c.depth].Next, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
			}
		}
		c.undoMove()
		if c.updateValue() {
			break
		}
	}
	if c.states[c.depth].Value == -1 {
		c.solvedMove[c.turn][c.board] = c.states[c.depth].Moves[0]
	}
	if c.config.EnablePrint {
		c.print("final res", c.states[c.depth].Value, printconfig.PrintConfig{Move: c.solvedMove[c.turn][c.board]})
	}
}

func (c *Core) staleMate() bool {
	if len(c.states[c.depth].Moves) != 0 {
		return false
	}
	c.states[c.depth].Value = 0
	if c.config.EnablePrint {
		c.print("stalemate", c.states[c.depth].Value, printconfig.PrintConfig{})
	}
	return true
}

func (c *Core) deadKing() bool {
	if !c.states[c.depth].Moves[c.states[c.depth].MoveIndex].IsKing() {
		return false
	}
	c.states[c.depth].Value = 1
	c.solvedMove[c.turn][c.board] = c.states[c.depth].Moves[c.states[c.depth].MoveIndex]
	if c.config.EnablePrint {
		c.print("dead king", c.states[c.depth].Value, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
	}
	return true
}

func (c *Core) doMove() {
	if c.config.EnablePrint {
		c.print("before move", c.states[c.depth].Value, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
	}
	c.states[c.depth].What = c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToY()]
	c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToY()] =
		c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromY()]
	c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromY()] = ' '
}

func (c *Core) undoMove() {
	c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].FromY()] =
		c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToY()]
	c.board[c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToX()][c.states[c.depth].Moves[c.states[c.depth].MoveIndex].ToY()] = c.states[c.depth].What
}

func (c *Core) updateValue() bool {
	if c.states[c.depth].Next <= c.states[c.depth].Value {
		return false
	}
	c.states[c.depth].Value = c.states[c.depth].Next
	c.solvedMove[c.turn][c.board] = c.states[c.depth].Moves[c.states[c.depth].MoveIndex]
	if c.config.EnablePrint {
		c.print("updated res", c.states[c.depth].Value, printconfig.PrintConfig{Move: c.states[c.depth].Moves[c.states[c.depth].MoveIndex]})
	}
	return c.states[c.depth].Value == 1
}

func (c *Core) show() {
	c.config.MaxPrintDepth = 0
	res := int8(123)
	c.print("show", res, printconfig.PrintConfig{})
	visited := []map[[4][4]byte]interface{}{{}, {}}
	c.depth = 0
	c.states = nil
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
		mv := c.solvedMove[c.turn][c.board]
		if mv == 0 {
			break
		}
		c.states = append(c.states, &state.State{Value: res, Moves: []move.Move{mv}})
		c.doMove()
		c.depth++
		res = c.solved[c.turn][c.board]
		c.turn = (c.turn + 1) % 2
		c.print("after move", res, printconfig.PrintConfig{Move: mv})
	}
}

// Play plays a game agains the solution.
func (c *Core) Play() {
	c.config.MaxPrintDepth = 0
	res := int8(123)
	c.print("play", res, printconfig.PrintConfig{})
	c.turn = 0
	visited := []map[[4][4]byte]interface{}{{}, {}}
	c.depth = 0
	c.states = nil
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
		mv := c.solvedMove[c.turn][c.board]
		if mv == 0 {
			break
		}
		c.states = append(c.states, &state.State{Value: res, Moves: []move.Move{mv}})
		c.doMove()
		c.depth++
		res = c.solved[c.turn][c.board]
		c.print("after move", res, printconfig.PrintConfig{Move: mv})

		fmt.Printf("> ")
		var fx, fy, tx, ty int
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
	}
}
