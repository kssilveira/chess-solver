// Package printconfig contains print configuration.
package printconfig

import "github.com/kssilveira/chess-solver/move"

// PrintConfig contains print configuration.
type PrintConfig struct {
	Move          move.Move
	ClearTerminal bool
}
