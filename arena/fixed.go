package arena

type Fixed struct {
	Linked
	pages [][]byte
	size  uint16
}

func (a *Fixed) alloc() (uint64, uint16) {
	pid := a.len()
	if pid&63 == 0 {
		a.bitset0 = append(a.bitset0, 0)
	}
	a.bitset1 = append(a.bitset1, new(bitset256))
	a.bitset2 = append(a.bitset2, new(bitset16k))
	a.pages = append(a.pages, make([]byte, a.size*pageGranules))
	return pid, 0
}

func (a *Fixed) write(p []byte) uint64 {
	return 0
}

func (a *Fixed) len() uint64 { return uint64(len(a.pages)) }
