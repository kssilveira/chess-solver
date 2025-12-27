// Package core constains the core logic.
package core

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"sync"
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
	PrintDepth    bool
	NumSolvers    int8
}

// PrintConfig contains print configuration.
type PrintConfig struct {
	Move          Move
	ClearTerminal bool
}

// Core contains the core logic.
type Core struct {
	config          Config
	clearTerminal   string
	writer          io.Writer
	visitedMutex    sync.RWMutex
	visited         []map[[4][4]byte]interface{}
	solvedMutex     sync.RWMutex
	solved          []map[[4][4]byte]int8
	solvedMoveMutex sync.RWMutex
	solvedMove      []map[[4][4]byte]Move
	solvers         []*Solver
}

// Solver contains one solver.
type Solver struct {
	core   *Core
	index  int8
	board  [4][4]byte
	turn   int8
	depth  int
	states []*State
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
	board := [4][4]byte{
		[4]byte([]byte("bnrk")),
		[4]byte([]byte("   p")),
		[4]byte([]byte("P   ")),
		[4]byte([]byte("KRNB")),
	}
	if config.NumSolvers == 0 {
		config.NumSolvers = 1
	}
	res := &Core{writer: writer, config: config,
		visited: []map[[4][4]byte]interface{}{{}, {}},
		solved:  []map[[4][4]byte]int8{{}, {}}, solvedMove: []map[[4][4]byte]Move{{}, {}},
		clearTerminal: "\033[H\033[2J"}
	if len(config.Board) > 1 {
		for i, row := range config.Board {
			board[i] = [4]byte([]byte(row))
		}
	}
	for i := int8(0); i < config.NumSolvers; i++ {
		solver := &Solver{core: res, index: i, board: board}
		res.solvers = append(res.solvers, solver)
	}
	return res
}

// Solve solves the board.
func (c *Core) Solve() {
	if c.config.NumSolvers == 1 {
		c.solvers[0].Solve()
	} else {
		var wg sync.WaitGroup
		for _, solver := range c.solvers {
			wg.Go(func() {
				solver.Solve()
			})
		}
		wg.Wait()
	}
	if c.config.EnablePrint && c.config.EnableShow {
		c.solvers[0].show()
	}
}

// Play plays a game agains the solution.
func (c *Core) Play() {
	c.solvers[0].Play()
}

func (c *Core) CanVisit(turn int8, board [4][4]byte) bool {
	if c.config.NumSolvers > 1 {
		c.visitedMutex.Lock()
		defer c.visitedMutex.Unlock()
	}
	_, ok := c.visited[turn][board]
	if ok {
		return false
	}
	c.visited[turn][board] = struct{}{}
	return true
}

func (c *Core) GetSolved(turn int8, board [4][4]byte) (int8, bool) {
	if c.config.NumSolvers > 1 {
		c.solvedMutex.Lock()
		defer c.solvedMutex.Unlock()
	}
	res, ok := c.solved[turn][board]
	return res, ok
}

func (c *Core) SetSolved(turn int8, board [4][4]byte, value int8) {
	if c.config.NumSolvers > 1 {
		c.solvedMutex.Lock()
		defer c.solvedMutex.Unlock()
	}
	c.solved[turn][board] = value
}

func (c *Core) GetSolvedMove(turn int8, board [4][4]byte) Move {
	if c.config.NumSolvers > 1 {
		c.solvedMoveMutex.Lock()
		defer c.solvedMoveMutex.Unlock()
	}
	return c.solvedMove[turn][board]
}

func (c *Core) SetSolvedMove(turn int8, board [4][4]byte, value Move) {
	if c.config.NumSolvers > 1 {
		c.solvedMoveMutex.Lock()
		defer c.solvedMoveMutex.Unlock()
	}
	c.solvedMove[turn][board] = value
}

func (s *Solver) Solve() *State {
	if s.core.config.PrintDepth && s.depth%100000 == 0 {
		fmt.Fprintf(s.core.writer, "%d\n", s.depth)
	}
	if len(s.states) == s.depth {
		s.states = append(s.states, &State{})
	}
	state := s.states[s.depth]
	if s.core.config.EnablePrint {
		s.print("after move", -1, PrintConfig{ClearTerminal: true})
	}
	if s.core.config.MaxDepth >= 0 && s.depth >= s.core.config.MaxDepth {
		state.Value = 0
		if s.core.config.EnablePrint {
			s.print("max depth", state.Value, PrintConfig{})
		}
		return state
	}
	if cap(state.Moves) == 0 {
		state.Moves = make([]Move, 0, 10)
	} else {
		state.Moves = state.Moves[:0]
	}
	s.moves(&state.Moves)
	sort(&state.Moves)
	s.move(state)
	return state
}

func (s *Solver) print(message string, res int8, cfg PrintConfig) {
	if s.core.config.MaxPrintDepth != 0 && s.depth > s.core.config.MaxPrintDepth {
		return
	}
	fmt.Fprintf(s.core.writer, "\n%s\n", message)
	fmt.Fprintf(s.core.writer, "turn: %d\n", s.turn)
	fmt.Fprintf(s.core.writer, "depth: %d\n", s.depth)
	fmt.Fprintf(s.core.writer, "res: %d\n", res)
	if cfg.Move != 0 {
		fx, fy, tx, ty := cfg.Move.Get()
		fmt.Fprintf(s.core.writer, "move: %s (%d, %d) => %s (%d, %d)\n", s.what(fx, fy), fx, fy, s.what(tx, ty), tx, ty)
	}
	fmt.Fprintln(s.core.writer, "______")
	fmt.Fprintln(s.core.writer, "|"+string(bytes.Join(toBytes(s.board), []byte("|\n|")))+"|")
	fmt.Fprintln(s.core.writer, "‾‾‾‾‾‾")
	if cfg.ClearTerminal {
		time.Sleep(s.core.config.SleepDuration)
		fmt.Fprint(s.core.writer, s.core.clearTerminal)
	}
}

func (s *Solver) what(x, y int8) string {
	return string(s.board[x][y])
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

func (s *Solver) moves(moves *[]Move) {
	for i := int8(0); i < 4; i++ {
		for j := int8(0); j < 4; j++ {
			piece := s.board[i][j]
			if colors[piece] != s.turn {
				continue
			}
			s.deltas(moves, i, j)
		}
	}
}

func (s *Solver) deltas(moves *[]Move, i, j int8) {
	piece := s.board[i][j]
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
		nextTurn := int8((s.turn + 1) % 2)
		if (kind == deltaDefault || kind == deltaOtherEmpty) && s.board[ni][nj] != ' ' && colors[s.board[ni][nj]] != nextTurn {
			continue
		}
		if kind == deltaEmpty && s.board[ni][nj] != ' ' {
			continue
		}
		if kind == deltaEnemy && colors[s.board[ni][nj]] != nextTurn {
			continue
		}
		if kind == deltaOtherEmpty && s.board[i+delta[3]][j+delta[4]] != ' ' {
			continue
		}
		*moves = append(*moves, NewMove(Move(i), Move(j), Move(ni), Move(nj), s.board[ni][nj] == 'k' || s.board[ni][nj] == 'K', s.board[ni][nj] != ' '))
	}
}

func sort(moves *[]Move) {
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

func (s *Solver) move(state *State) {
	state.Value = -1
	if len(state.Moves) == 0 {
		state.Value = 0
		if s.core.config.EnablePrint {
			s.print("stalemate", state.Value, PrintConfig{})
		}
		return
	}
	for state.MoveIndex = 0; state.MoveIndex < int8(len(state.Moves)); state.MoveIndex++ {
		if state.Moves[state.MoveIndex].IsKing() {
			state.Value = 1
			s.core.SetSolvedMove(s.turn, s.board, state.Moves[state.MoveIndex])
			if s.core.config.EnablePrint {
				s.print("dead king", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			return
		}
		if s.core.config.EnablePrint {
			s.print("before move", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
		}
		state.What = s.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		s.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = s.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()]
		s.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = ' '
		state.Next = 0
		ok := false
		if state.Next, ok = s.core.GetSolved(s.turn, s.board); ok {
			if s.core.config.EnablePrint {
				s.print("solved[]", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
		} else if s.core.CanVisit(s.turn, s.board) {
			s.turn = int8((s.turn + 1) % 2)
			s.depth++
			nextState := s.Solve()
			state.Next = -nextState.Value
			s.depth--
			s.turn = int8((s.turn + 1) % 2)
			if s.core.config.EnablePrint {
				s.print("solve()", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			s.core.SetSolved(s.turn, s.board, state.Next)
		} else {
			if s.core.config.EnablePrint {
				s.print("repeated", state.Next, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
		}
		s.board[state.Moves[state.MoveIndex].FromX()][state.Moves[state.MoveIndex].FromY()] = s.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()]
		s.board[state.Moves[state.MoveIndex].ToX()][state.Moves[state.MoveIndex].ToY()] = state.What
		if state.Next > state.Value {
			state.Value = state.Next
			s.core.SetSolvedMove(s.turn, s.board, state.Moves[state.MoveIndex])
			if s.core.config.EnablePrint {
				s.print("updated res", state.Value, PrintConfig{Move: state.Moves[state.MoveIndex]})
			}
			if state.Value == 1 {
				break
			}
		}
	}
	if state.Value == -1 {
		s.core.SetSolvedMove(s.turn, s.board, state.Moves[0])
	}
	if s.core.config.EnablePrint {
		s.print("final res", state.Value, PrintConfig{Move: s.core.GetSolvedMove(s.turn, s.board)})
	}
}

func (s *Solver) show() {
	s.core.config.MaxPrintDepth = 0
	res := int8(123)
	s.print("show", res, PrintConfig{})
	visited := []map[[4][4]byte]interface{}{{}, {}}
	for {
		if _, ok := visited[s.turn][s.board]; ok {
			break
		}
		visited[s.turn][s.board] = true
		move := s.core.solvedMove[s.turn][s.board]
		if move == 0 {
			break
		}
		s.print("before move", res, PrintConfig{Move: move})

		fx, fy, tx, ty := move.Get()
		s.board[tx][ty] = s.board[fx][fy]
		s.board[fx][fy] = ' '
		s.depth++
		res = s.core.solved[s.turn][s.board]
		s.turn = (s.turn + 1) % 2
		s.print("after move", res, PrintConfig{Move: move})
	}
}

// Play plays a game agains the solution.
func (s *Solver) Play() {
	s.core.config.MaxPrintDepth = 0
	res := int8(123)
	s.print("play", res, PrintConfig{})
	s.turn = 0
	visited := []map[[4][4]byte]interface{}{{}, {}}
	for {
		if _, ok := visited[s.turn][s.board]; ok {
			break
		}
		visited[s.turn][s.board] = true
		move := s.core.solvedMove[s.turn][s.board]
		if move == 0 {
			break
		}
		s.print("before move", res, PrintConfig{Move: move})

		fx, fy, tx, ty := move.Get()
		s.board[tx][ty] = s.board[fx][fy]
		s.board[fx][fy] = ' '
		s.depth++
		res = s.core.solved[s.turn][s.board]
		s.print("after move", res, PrintConfig{Move: move})

		fmt.Printf("> ")
		fmt.Scanf("%d%d%d%d", &fx, &fy, &tx, &ty)
		s.board[tx][ty] = s.board[fx][fy]
		s.board[fx][fy] = ' '
	}
}
