package invalid

type shiftState struct {
	lastWriteCycle uint64
	lastWriteValid bool
}

func (m *shiftState) writeShiftBit(cycle uint64, ok bool) {
	m.lastWriteCycle, m.lastWriteValid = cycle, ok
}
