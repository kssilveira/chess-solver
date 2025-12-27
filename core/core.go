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
)

// Core contains the core logic.
type Core struct {
	writer        io.Writer
	config        config.Config
	board         [4][4]byte
	turn          int
	clearTerminal string
	visited       []map[[4][4]byte]interface{}
	solved        []map[[4][4]byte]int
	solvedMove    []map[[4][4]byte]move.Move
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
func New(writer io.Writer, config config.Config) *Core {
	res := &Core{writer: writer, config: config, board: [4][4]byte{
		[4]byte([]byte("bnrk")),
		[4]byte([]byte("   p")),
		[4]byte([]byte("P   ")),
		[4]byte([]byte("KRNB")),
	},
		visited: []map[[4][4]byte]interface{}{{}, {}},
		solved:  []map[[4][4]byte]int{{}, {}}, solvedMove: []map[[4][4]byte]move.Move{{}, {}},
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

func (c *Core) solve() int {
	if c.config.PrintDepth && c.depth%100000 == 0 {
		fmt.Fprintf(c.writer, "%d\n", c.depth)
	}
	if c.config.EnablePrint {
		c.print("after move", -1, printconfig.PrintConfig{ClearTerminal: true})
	}
	if c.config.MaxDepth >= 0 && c.depth >= c.config.MaxDepth {
		res := 0
		if c.config.EnablePrint {
			c.print("max depth", res, printconfig.PrintConfig{})
		}
		return res
	}
	moves := make([]move.Move, 0, 10)
	c.moves(&moves)
	c.sort(moves)
	return c.move(moves)
}

func (c *Core) print(message string, res int, cfg printconfig.PrintConfig) {
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

func (c *Core) what(x, y int) string {
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
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != c.turn {
				continue
			}
			c.deltas(moves, i, j)
		}
	}
}

func (c *Core) deltas(moves *[]move.Move, i, j int) {
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
		nextTurn := (c.turn + 1) % 2
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
		*moves = append(
			*moves,
			move.NewMove(
				move.Move(i), move.Move(j), move.Move(ni), move.Move(nj),
				c.board[ni][nj] == 'k' || c.board[ni][nj] == 'K',
				c.board[ni][nj] != ' '))
	}
}

func (c *Core) sort(moves []move.Move) {
	slices.SortFunc(moves, func(i, j move.Move) int {
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

func (c *Core) move(moves []move.Move) int {
	res := -1
	if res, ok := c.staleMate(moves); ok {
		return res
	}
	for _, move := range moves {
		if res, ok := c.deadKing(move); ok {
			return res
		}
		what := c.doMove(move, res)
		next := 0
		ok := false
		if next, ok = c.solved[c.turn][c.board]; ok {
			if c.config.EnablePrint {
				c.print("solved[]", next, printconfig.PrintConfig{Move: move})
			}
		} else if _, ok = c.visited[c.turn][c.board]; !ok {
			c.visited[c.turn][c.board] = struct{}{}
			c.turn = (c.turn + 1) % 2
			c.depth++
			next = -c.solve()
			c.depth--
			c.turn = (c.turn + 1) % 2
			c.solved[c.turn][c.board] = next
			if c.config.EnablePrint {
				c.print("solve()", next, printconfig.PrintConfig{Move: move})
			}
		} else {
			if c.config.EnablePrint {
				c.print("repeated", next, printconfig.PrintConfig{Move: move})
			}
		}
		c.undoMove(move, what)
		if c.updateValue(&res, next, move) {
			break
		}
	}
	if res == -1 {
		c.solvedMove[c.turn][c.board] = moves[0]
	}
	if c.config.EnablePrint {
		c.print("final res", res, printconfig.PrintConfig{Move: c.solvedMove[c.turn][c.board]})
	}
	return res
}

func (c *Core) staleMate(moves []move.Move) (int, bool) {
	if len(moves) != 0 {
		return 0, false
	}
	res := 0
	if c.config.EnablePrint {
		c.print("stalemate", res, printconfig.PrintConfig{})
	}
	return res, true
}

func (c *Core) deadKing(move move.Move) (int, bool) {
	if !move.IsKing() {
		return 0, false
	}
	res := 1
	c.solvedMove[c.turn][c.board] = move
	if c.config.EnablePrint {
		c.print("dead king", res, printconfig.PrintConfig{Move: move})
	}
	return res, true
}

func (c *Core) doMove(move move.Move, res int) byte {
	if c.config.EnablePrint {
		c.print("before move", res, printconfig.PrintConfig{Move: move})
	}
	what := c.board[move.ToX()][move.ToY()]
	c.board[move.ToX()][move.ToY()] =
		c.board[move.FromX()][move.FromY()]
	c.board[move.FromX()][move.FromY()] = ' '
	return what
}

func (c *Core) undoMove(move move.Move, what byte) {
	c.board[move.FromX()][move.FromY()] =
		c.board[move.ToX()][move.ToY()]
	c.board[move.ToX()][move.ToY()] = what
}

func (c *Core) updateValue(res *int, next int, move move.Move) bool {
	if next <= *res {
		return false
	}
	*res = next
	c.solvedMove[c.turn][c.board] = move
	if c.config.EnablePrint {
		c.print("updated res", *res, printconfig.PrintConfig{Move: move})
	}
	return *res == 1
}

func (c *Core) show() {
	c.config.MaxPrintDepth = 0
	res := 123
	c.print("show", res, printconfig.PrintConfig{})
	visited := []map[[4][4]byte]interface{}{{}, {}}
	c.depth = 0
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
		move := c.solvedMove[c.turn][c.board]
		if move == 0 {
			break
		}
		c.doMove(move, res)
		c.depth++
		res = c.solved[c.turn][c.board]
		c.turn = (c.turn + 1) % 2
		c.print("after move", res, printconfig.PrintConfig{Move: move})
	}
}

// Play plays a game agains the solution.
func (c *Core) Play() {
	c.config.MaxPrintDepth = 0
	res := 123
	c.print("play", res, printconfig.PrintConfig{})
	c.turn = 0
	visited := []map[[4][4]byte]interface{}{{}, {}}
	c.depth = 0
	for {
		if _, ok := visited[c.turn][c.board]; ok {
			break
		}
		visited[c.turn][c.board] = true
		move := c.solvedMove[c.turn][c.board]
		if move == 0 {
			break
		}
		c.doMove(move, res)
		c.depth++
		res = c.solved[c.turn][c.board]
		c.print("after move", res, printconfig.PrintConfig{Move: move})

		fmt.Printf("> ")
		var fx, fy, tx, ty int
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
	}
}
