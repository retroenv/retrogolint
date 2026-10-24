package valid

type shiftState struct {
	lastWriteCycle uint64
	lastWriteValid bool
}

func (m *shiftState) writeShiftBit(cycle uint64, ok bool) {
	m.lastWriteCycle = cycle
	m.lastWriteValid = ok
}

func (m *shiftState) swapState(other *shiftState) {
	m.lastWriteCycle, other.lastWriteCycle = other.lastWriteCycle, m.lastWriteCycle
}
