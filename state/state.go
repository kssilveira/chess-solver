// Package state contains the recursion state.
package state

import "github.com/kssilveira/chess-solver/move"

// State contains the recursion state.
type State struct {
	Value     int8
	Next      int8
	MoveIndex int8
	What      byte
	OK        bool
	Moves     []move.Move
}
