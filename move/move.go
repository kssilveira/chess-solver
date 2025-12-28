// Package move contains move logic.
package move

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
func (m Move) Get() (int, int, int, int) {
	return m.FromX(), m.FromY(), m.ToX(), m.ToY()
}

// FromX returns from x.
func (m Move) FromX() int {
	return int(m & 0b11)
}

// FromY returns from y.
func (m Move) FromY() int {
	return int((m & 0b1100) >> 2)
}

// ToX returns to x.
func (m Move) ToX() int {
	return int((m & 0b110000) >> 4)
}

// ToY returns to y.
func (m Move) ToY() int {
	return int((m & 0b11000000) >> 6)
}

// IsKing returns is king.
func (m Move) IsKing() bool {
	return m&(1<<8) != 0
}

// IsCapture returns is capture.
func (m Move) IsCapture() bool {
	return m&(1<<9) != 0
}

// SetNextIsKing set next is king.
func (m *Move) SetNextIsKing() {
	*m |= 1 << 10
}

// NextIsKing returns next is king.
func (m Move) NextIsKing() bool {
	return m&(1<<10) != 0
}

// SetNextIsCapture sets next is capture.
func (m *Move) SetNextIsCapture() {
	*m |= 1 << 11
}

// NextIsCapture returns next is capture.
func (m Move) NextIsCapture() bool {
	return m&(1<<11) != 0
}
