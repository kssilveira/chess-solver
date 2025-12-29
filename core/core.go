// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
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
	board         [6][4]byte
	visited       []map[[6][4]byte]interface{}
	solved        []map[[6][4]byte]int
	solvedMove    []map[[6][4]byte]move.Move
	sharedMoves   []move.Move
}

const (
	deltaDefault = iota
	deltaEmpty
	deltaEnemy
	deltaOtherEmpty
)

var (
	colors = map[byte]int{
		' ': -1,
		'P': 0, 'K': 0, 'R': 0, 'N': 0, 'B': 0, 'X': 0,
		'p': 1, 'k': 1, 'r': 1, 'n': 1, 'b': 1, 'x': 1,
	}
	deltas = map[byte][][]int{
		'P': {{-1, 0, deltaEmpty}, {-1, -1, deltaEnemy}, {-1, 1, deltaEnemy}},
		'R': {{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		'B': {{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		'K': {{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		'N': {
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
		'p': {{1, 0, deltaEmpty}, {1, -1, deltaEnemy}, {1, 1, deltaEnemy}},
		'r': {{-1, 0}, {1, 0}, {0, -1}, {0, 1}},
		'b': {{-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		'k': {{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, 1}, {1, -1}, {-1, 1}},
		'n': {
			{-2, -1, deltaOtherEmpty, -1, 0}, {-2, 1, deltaOtherEmpty, -1, 0},
			{-1, -2, deltaOtherEmpty, 0, -1}, {1, -2, deltaOtherEmpty, 0, -1},
			{2, -1, deltaOtherEmpty, 1, 0}, {2, 1, deltaOtherEmpty, 1, 0},
			{-1, 2, deltaOtherEmpty, 0, 1}, {1, 2, deltaOtherEmpty, 0, 1},
		},
	}
	visitedStruct = struct{}{}
	promos        = map[byte][]byte{
		'P': {'R', 'B', 'N'},
		'p': {'r', 'b', 'n'},
	}
	undoPromos = map[byte]byte{
		'R': 'P', 'B': 'P', 'N': 'P',
		'r': 'p', 'b': 'p', 'n': 'p',
	}
	deadX = map[byte]int{
		'R': 5, 'B': 5, 'N': 5, 'P': 5,
		'r': 4, 'b': 4, 'n': 4, 'p': 4,
	}
	deadY = map[byte]int{
		'R': 0, 'B': 1, 'N': 2, 'P': 3,
		'r': 0, 'b': 1, 'n': 2, 'p': 3,
	}
	deadXY = [][]byte{
		[]byte("RBNP"),
		[]byte("rbnp"),
	}
)

// New creates a new core.
func New(writer io.Writer, config config.Config) *Core {
	res := &Core{
		writer: writer, config: config,
		board: [6][4]byte{
			[4]byte([]byte("bnrk")),
			[4]byte([]byte("   p")),
			[4]byte([]byte("P   ")),
			[4]byte([]byte("KRNB")),
			[4]byte([]byte("0000")),
			[4]byte([]byte("0000")),
		},
		visited: []map[[6][4]byte]interface{}{
			make(map[[6][4]byte]interface{}, 100000),
			make(map[[6][4]byte]interface{}, 100000),
		},
		solved: []map[[6][4]byte]int{
			make(map[[6][4]byte]int, 100000),
			make(map[[6][4]byte]int, 100000),
		},
		solvedMove: []map[[6][4]byte]move.Move{
			make(map[[6][4]byte]move.Move, 100000),
			make(map[[6][4]byte]move.Move, 100000),
		},
		sharedMoves:   make([]move.Move, 0, 15),
		clearTerminal: "\033[H\033[2J"}
	if len(config.Board) > 1 {
		for i, row := range strings.Split(config.Board, ",") {
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
	Moves    [100]move.Move
	NumMoves int
	Move     move.Move
	Value    int
	Next     int
	Index    int
	What     byte
}

func (c *Core) solve() (int, int) {
	stack := make([]State, 0, 100000)
	c.call(&stack)
	overall := -1
	maxDepth := 0
	maxVisited := 0
	for len(stack) > 0 {
		depth := len(stack) - 1
		turn := depth % 2
		state := &stack[depth]
		c.updateMaxDepth(&maxDepth, depth)
		c.updateMaxVisited(&maxVisited)
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
			c.solvedMove[(turn+1)%2][c.board] = state.Moves[0]
		}
		c.print("final res", state.Value, depth, turn, printconfig.PrintConfig{Move: c.solvedMove[(turn+1)%2][c.board]})
		overall = c.doReturn(&stack)
	}
	return overall, maxDepth + 1
}

func (c *Core) updateMaxDepth(maxDepth *int, depth int) {
	if depth > *maxDepth {
		*maxDepth = depth
		if c.config.PrintDepth && *maxDepth%1000000 == 0 {
			fmt.Fprintf(c.writer, "depth: %d\n", *maxDepth)
		}
	}
}

func (c *Core) updateMaxVisited(maxVisited *int) {
	numVisited := len(c.visited[0])
	if numVisited > *maxVisited {
		*maxVisited = numVisited
		if c.config.PrintDepth && *maxVisited%10000000 == 0 {
			fmt.Fprintf(c.writer, "visited: %d\n", *maxVisited)
		}
	}
}

func (c *Core) call(stack *[]State) {
	*stack = append(*stack, State{Value: -1})
	depth := len(*stack) - 1
	turn := depth % 2
	state := &(*stack)[depth]

	c.sharedMoves = c.sharedMoves[:0]
	c.moves(&c.sharedMoves, turn)
	for i, move := range c.sharedMoves {
		state.Moves[i] = move
	}
	state.NumMoves = len(c.sharedMoves)
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

func toBytes(board [6][4]byte) [][]byte {
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
	if !c.config.EnableDrop {
		c.sort(*moves)
		return
	}
	for index := 0; index < 4; index++ {
		value := c.board[4+turn][index]
		if value == '0' {
			continue
		}
		piece := deadXY[turn][index]
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				if c.board[i][j] != ' ' {
					continue
				}
				if c.config.EnablePromotion && ((i == 0 && piece == 'P') || (i == 3 && piece == 'p')) {
					continue
				}
				move := move.NewMove(move.Move(turn), move.Move(index), move.Move(i), move.Move(j), false, false)
				move.SetDrop()
				*moves = append(*moves, move)
			}
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
		move := move.NewMove(
			move.Move(i), move.Move(j), move.Move(ni), move.Move(nj),
			c.board[ni][nj] == 'k' || c.board[ni][nj] == 'K',
			c.board[ni][nj] != ' ')
		if c.config.EnablePromotion && ((ni == 0 && piece == 'P') || (ni == 3 && piece == 'p')) {
			move.SetPromotion(1)
			*moves = append(*moves, move)
			move.SetPromotion(2)
			*moves = append(*moves, move)
			move.SetPromotion(3)
		}
		*moves = append(*moves, move)
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
		if i.IsDrop() {
			return -1
		}
		if j.IsDrop() {
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
	c.solvedMove[(turn+1)%2][c.board] = move
	c.print("dead king", res, depth, turn, printconfig.PrintConfig{Move: move})
	return res, true
}

func (c *Core) doMove(move move.Move, res, depth, turn int) byte {
	c.print("before move", res, depth, turn, printconfig.PrintConfig{Move: move})
	what := c.board[move.ToX()][move.ToY()]
	from := c.board[move.FromX()][move.FromY()]
	if move.IsDrop() {
		from = deadXY[move.FromX()][move.FromY()]
		c.board[4+move.FromX()][move.FromY()]--
	} else {
		c.board[move.FromX()][move.FromY()] = ' '
	}
	c.board[move.ToX()][move.ToY()] = from
	if move.IsCapture() && !move.IsKing() && what != 'x' && what != 'X' {
		c.board[deadX[what]][deadY[what]]++
	}
	promotion := move.Promotion()
	if promotion == 0 {
		return what
	}
	c.board[move.ToX()][move.ToY()] = promos[c.board[move.ToX()][move.ToY()]][promotion-1]
	return what
}

func (c *Core) undoMove(move move.Move, what byte) {
	if move.IsDrop() {
		c.board[4+move.FromX()][move.FromY()]++
	} else {
		c.board[move.FromX()][move.FromY()] = c.board[move.ToX()][move.ToY()]
	}
	c.board[move.ToX()][move.ToY()] = what
	if move.IsCapture() && !move.IsKing() && what != 'x' && what != 'X' {
		c.board[deadX[what]][deadY[what]]--
	}
	promotion := move.Promotion()
	if promotion == 0 {
		return
	}
	c.board[move.FromX()][move.FromY()] = undoPromos[c.board[move.FromX()][move.FromY()]]
}

func (c *Core) updateValue(res *int, next int, move move.Move, depth, turn int) bool {
	if next <= *res {
		return false
	}
	*res = next
	c.solvedMove[(turn+1)%2][c.board] = move
	c.print("updated res", *res, depth, turn, printconfig.PrintConfig{Move: move})
	return *res == 1
}

func (c *Core) show() {
	c.config.MaxPrintDepth = 0
	res := 123
	visited := []map[[6][4]byte]interface{}{{}, {}}
	depth := 0
	turn := 0
	c.print("show", res, depth, turn, printconfig.PrintConfig{})
	for {
		if _, ok := visited[turn][c.board]; ok {
			break
		}
		visited[turn][c.board] = true
		move := c.solvedMove[(turn+1)%2][c.board]
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
	visited := []map[[6][4]byte]interface{}{{}, {}}
	depth := 0
	c.print("play", res, depth, turn, printconfig.PrintConfig{})
	for {
		if _, ok := visited[turn][c.board]; ok {
			break
		}
		visited[turn][c.board] = true
		move := c.solvedMove[(turn+1)%2][c.board]
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

// RunAll runs all configs.
func RunAll(writer io.Writer, configs []config.Config) {
	buffers := []*bytes.Buffer{}
	var wg sync.WaitGroup
	for _, config := range configs {
		var buffer bytes.Buffer
		buffers = append(buffers, &buffer)
		wg.Go(func() {
			config.MaxPrintDepth = -1
			core := New(&buffer, config)
			core.Solve()
		})
	}
	wg.Wait()
	for i, config := range configs {
		desc := []string{
			fmt.Sprintf("--board='%s'", config.Board),
		}
		if config.EnablePromotion {
			desc = append(desc, "--enable_promotion")
		}
		if config.EnableDrop {
			desc = append(desc, "--enable_drop")
		}
		fmt.Fprintf(writer, "\n%s\n%s", strings.Join(desc, " "), buffers[i].String())
	}
}
