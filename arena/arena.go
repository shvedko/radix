package arena

import (
	"math"
	"math/bits"
)

const (
	granuleBytes = 8
	pageGranules = 16384
	pageBytes    = pageGranules * granuleBytes
	_            = pageBytes / 1024
	_            = math.MaxUint64 / pageBytes
)

type granule [granuleBytes]byte
type page [pageGranules]granule
type bitset16k [pageGranules / 64]uint64
type bitset256 [pageGranules / 64 / 64]uint64

type hint struct {
	pid uint64
	gid uint16
}

type Linked struct {
	bitset0 []uint64
	bitset1 []*bitset256
	bitset2 []*bitset16k
	pages   []*page
	hint
}

func pack(pid uint64, gid uint16) uint64 { return (pid << 14) | uint64(gid) }
func unpack(id uint64) (uint64, uint16)  { return id >> 14, uint16(id & 0x3FFF) }

func (a *Linked) next(pid uint64, gid uint16) (uint64, uint16) {
	if pid >= uint64(len(a.pages)) {
		return a.alloc()
	}

	pid, gid, ok := a.scan(pid, gid)
	if ok {
		return pid, gid
	}

	return a.alloc()
}

func (a *Linked) scan(pid uint64, gid uint16) (uint64, uint16, bool) {
	gid, ok := a.find(pid, gid)
	if ok {
		return pid, gid, true
	}

	for i := (pid + 1) >> 6; i < uint64(len(a.bitset0)); i++ {
		mask := a.bitset0[i]
		if mask != ^uint64(0) {
			pid = i<<6 + uint64(bits.TrailingZeros64(^mask))
			gid, ok = a.find(pid, 0)
			if ok {
				return pid, gid, true
			}
		}
	}

	return 0, 0, false
}

func (a *Linked) find(pid uint64, gid uint16) (uint16, bool) {
	gid1 := gid >> 12
	gid2 := gid >> 6

	for i := gid1; i < 4; i++ {
		mask1 := a.bitset1[pid][i]
		if i == gid1 {
			mask1 |= 1<<(gid2&63) - 1
		}
		if mask1 == ^uint64(0) {
			continue
		}
		idx2 := i<<6 + uint16(bits.TrailingZeros64(^mask1))
		mask2 := a.bitset2[pid][idx2]
		if idx2 == gid2 {
			mask2 |= 1<<(gid&63) - 1
		}

		return idx2<<6 + uint16(bits.TrailingZeros64(^mask2)), true
	}

	return 0, false
}

func (a *Linked) mark(pid uint64, gid uint16, occupied bool) {
	idx2 := gid >> 6
	bit2 := uint64(1) << (gid & 63)
	idx1 := idx2 >> 6
	bit1 := uint64(1) << (idx2 & 63)

	if occupied {
		a.bitset2[pid][idx2] |= bit2

		if a.bitset2[pid][idx2] == ^uint64(0) {
			a.bitset1[pid][idx1] |= bit1

			if a.bitset1[pid][0] == ^uint64(0) &&
				a.bitset1[pid][1] == ^uint64(0) &&
				a.bitset1[pid][2] == ^uint64(0) &&
				a.bitset1[pid][3] == ^uint64(0) {
				a.bitset0[pid>>6] |= 1 << (pid & 63)
			}
		}
	} else {
		a.bitset2[pid][idx2] &^= bit2
		a.bitset1[pid][idx1] &^= bit1
		a.bitset0[pid>>6] &^= 1 << (pid & 63)
	}
}

func (a *Linked) alloc() (uint64, uint16) {
	i, pid := 1, uint64(len(a.pages))
	switch {
	case pid >= 64:
		i = 8
	case pid >= 8:
		i = 2
	}

	pages := make([]page, i)
	bits1 := make([]bitset256, i)
	bits2 := make([]bitset16k, i)

	for i > 0 {
		i--
		if len(a.pages)%64 == 0 {
			a.bitset0 = append(a.bitset0, 0)
		}
		a.bitset1 = append(a.bitset1, &bits1[i])
		a.bitset2 = append(a.bitset2, &bits2[i])
		a.pages = append(a.pages, &pages[i])
	}

	return pid, 0
}
