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
	config        config.Config
	clearTerminal string
	writer        io.Writer
	board         [4][4]byte
	visited       []map[[4][4]byte]interface{}
	solved        []map[[4][4]byte]int
	solvedMove    []map[[4][4]byte]move.Move
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
		visited: []map[[4][4]byte]interface{}{
			make(map[[4][4]byte]interface{}, 100000),
			make(map[[4][4]byte]interface{}, 100000),
		},
		solved: []map[[4][4]byte]int{
			make(map[[4][4]byte]int, 100000),
			make(map[[4][4]byte]int, 100000),
		},
		solvedMove: []map[[4][4]byte]move.Move{
			make(map[[4][4]byte]move.Move, 100000),
			make(map[[4][4]byte]move.Move, 100000),
		},
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
	res, maxDepth := c.solve()
	fmt.Fprintf(c.writer, "\nmax depth: %d\n", maxDepth)
	fmt.Fprintf(c.writer, "overall res: %d\n", res)
	if c.config.EnableShow {
		c.show()
	}
}

// State contains the recursion state.
type State struct {
	Moves    [15]move.Move
	NumMoves int
	Move     move.Move
	Value    int
	Next     int
	Index    int
	What     byte
}

var (
	visitedStruct = struct{}{}
)

func (c *Core) solve() (int, int) {
	stack := make([]State, 0, 100000)
	c.call(&stack)
	overall := -1
	maxDepth := 0
	for len(stack) > 0 {
		depth := len(stack) - 1
		turn := depth % 2
		state := &stack[depth]
		if depth > maxDepth {
			maxDepth = depth
			if c.config.PrintDepth && maxDepth%1000000 == 0 {
				fmt.Fprintf(c.writer, "max depth: %d\n", maxDepth)
			}
		}
		if state.Index == 0 {
			c.print("after move", -1, depth, turn, printconfig.PrintConfig{ClearTerminal: true})
		}
		if res, ok := c.staleMate(state.NumMoves, depth, turn); ok {
			state.Value = res
			overall = c.doReturn(&stack)
			continue
		}
		if state.Index < state.NumMoves {
			state.Move = state.Moves[state.Index]
			if res, ok := c.deadKing(state.Move, depth, turn); ok {
				state.Value = res
				overall = c.doReturn(&stack)
				continue
			}
			state.What = c.doMove(state.Move, state.Value, depth, turn)
			state.Next = 0
			ok := false
			if state.Next, ok = c.solved[turn][c.board]; ok {
				c.print("solved[]", state.Next, depth, turn, printconfig.PrintConfig{Move: state.Move})
			} else if _, ok = c.visited[turn][c.board]; ok {
				c.print("repeated", state.Next, depth, turn, printconfig.PrintConfig{Move: state.Move})
			} else {
				c.visited[turn][c.board] = visitedStruct
				c.call(&stack)
				continue
			}
			c.afterReturn(stack)
			continue
		}
		if state.Value == -1 {
			c.solvedMove[turn][c.board] = state.Moves[0]
		}
		c.print("final res", state.Value, depth, turn, printconfig.PrintConfig{Move: c.solvedMove[turn][c.board]})
		overall = c.doReturn(&stack)
	}
	return overall, maxDepth + 1
}

var (
	sharedMoves = make([]move.Move, 0, 15)
)

func (c *Core) call(stack *[]State) {
	*stack = append(*stack, State{Value: -1})
	depth := len(*stack) - 1
	turn := depth % 2
	state := &(*stack)[depth]

	sharedMoves = sharedMoves[:0]
	c.moves(&sharedMoves, turn)
	for i, move := range sharedMoves {
		state.Moves[i] = move
	}
	state.NumMoves = len(sharedMoves)
}

func (c *Core) doReturn(stack *[]State) int {
	depth := len(*stack) - 1
	state := &(*stack)[depth]

	next := -state.Value

	*stack = (*stack)[:depth]
	depth = len(*stack) - 1
	turn := depth % 2
	if depth < 0 {
		return state.Value
	}
	state = &(*stack)[depth]

	c.solved[turn][c.board] = next
	state.Next = next
	c.print("solve()", next, depth, turn, printconfig.PrintConfig{Move: state.Move})
	c.afterReturn(*stack)
	return state.Value
}

func (c *Core) afterReturn(stack []State) {
	depth := len(stack) - 1
	turn := depth % 2
	state := &stack[depth]
	c.undoMove(state.Move, state.What)
	if c.updateValue(&state.Value, state.Next, state.Move, depth, turn) {
		state.Index = state.NumMoves
	}
	state.Index++
}

func (c *Core) print(message string, value, depth, turn int, cfg printconfig.PrintConfig) {
	if c.config.MaxPrintDepth != 0 && depth > c.config.MaxPrintDepth {
		return
	}
	fmt.Fprintf(c.writer, "\n%s\n", message)
	fmt.Fprintf(c.writer, "turn: %d\n", turn)
	fmt.Fprintf(c.writer, "depth: %d\n", depth)
	fmt.Fprintf(c.writer, "res: %d\n", value)
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

func (c *Core) moves(moves *[]move.Move, turn int) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			piece := c.board[i][j]
			if colors[piece] != turn {
				continue
			}
			c.deltas(moves, turn, i, j)
		}
	}
	c.sort(*moves)
}

func (c *Core) deltas(moves *[]move.Move, turn, i, j int) {
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
		nextTurn := (turn + 1) % 2
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

func (c *Core) staleMate(moves, depth, turn int) (int, bool) {
	if moves != 0 {
		return 0, false
	}
	res := 0
	c.print("stalemate", res, depth, turn, printconfig.PrintConfig{})
	return res, true
}

func (c *Core) deadKing(move move.Move, depth, turn int) (int, bool) {
	if !move.IsKing() {
		return 0, false
	}
	res := 1
	c.solvedMove[turn][c.board] = move
	c.print("dead king", res, depth, turn, printconfig.PrintConfig{Move: move})
	return res, true
}

func (c *Core) doMove(move move.Move, res, depth, turn int) byte {
	c.print("before move", res, depth, turn, printconfig.PrintConfig{Move: move})
	what := c.board[move.ToX()][move.ToY()]
	c.board[move.ToX()][move.ToY()] = c.board[move.FromX()][move.FromY()]
	c.board[move.FromX()][move.FromY()] = ' '
	return what
}

func (c *Core) undoMove(move move.Move, what byte) {
	c.board[move.FromX()][move.FromY()] = c.board[move.ToX()][move.ToY()]
	c.board[move.ToX()][move.ToY()] = what
}

func (c *Core) updateValue(res *int, next int, move move.Move, depth, turn int) bool {
	if next <= *res {
		return false
	}
	*res = next
	c.solvedMove[turn][c.board] = move
	c.print("updated res", *res, depth, turn, printconfig.PrintConfig{Move: move})
	return *res == 1
}

func (c *Core) show() {
	c.config.MaxPrintDepth = 0
	res := 123
	visited := []map[[4][4]byte]interface{}{{}, {}}
	depth := 0
	turn := 0
	c.print("show", res, depth, turn, printconfig.PrintConfig{})
	for {
		if _, ok := visited[turn][c.board]; ok {
			break
		}
		visited[turn][c.board] = true
		move := c.solvedMove[turn][c.board]
		if move == 0 {
			break
		}
		c.doMove(move, res, depth, turn)
		res = c.solved[turn][c.board]
		depth++
		turn = (turn + 1) % 2
		c.print("after move", res, depth, turn, printconfig.PrintConfig{Move: move})
	}
}

// Play plays a game agains the solution.
func (c *Core) Play() {
	c.config.MaxPrintDepth = 0
	res := 123
	turn := 0
	visited := []map[[4][4]byte]interface{}{{}, {}}
	depth := 0
	c.print("play", res, depth, turn, printconfig.PrintConfig{})
	for {
		if _, ok := visited[turn][c.board]; ok {
			break
		}
		visited[turn][c.board] = true
		move := c.solvedMove[turn][c.board]
		if move == 0 {
			break
		}
		c.doMove(move, res, depth, turn)
		depth++
		res = c.solved[turn][c.board]
		c.print("after move", res, depth, turn, printconfig.PrintConfig{Move: move})

		fmt.Printf("> ")
		var fx, fy, tx, ty int
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		c.board[tx][ty] = c.board[fx][fy]
		c.board[fx][fy] = ' '
	}
}
