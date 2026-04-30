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
	//pid += uint64(gid >> 14)
	//gid &= 0x3FFF
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

func (a *Linked) write(p []byte) uint64 {
	pid, gid := a.next(a.pid, a.gid)
	rid := pack(pid, gid)

	for {
		a.mark(pid, gid, true)
		h := a.granule(pid, gid)

		if len(p) < 8 {
			h[0] = 0xF0 | byte(len(p)) // T.5
			copy(h[1:], p)

			return rid
		}

		size := min(129, len(p)/8)
		if size > 1 {
			size = a.need(size, pid, gid)
			if size > 1 {
				h[0] = byte(size - 2) // T.1
				copy(h[1:], p)
				p = p[7:]

				for i := 1; i < size; i++ {
					gid++
					if gid == pageGranules {
						pid++
						gid = 0
					}

					a.mark(pid, gid, true)
					h = a.granule(pid, gid)
					copy(h[:], p)
					p = p[8:]
				}

				gid++
				if gid == pageGranules {
					pid++
					gid = 0
				}
				continue
			}
		}

		pid1, gid1 := a.next(pid, gid)

		const (
			T0x80 = 0x3F + 1 + 1
			T0xC0 = 0xFFF + T0x80 + 1
			T0xD0 = 0x7FFFF + T0xC0 + 1
			T0xD8 = 0x7FFFFFF + T0xD0 + 1
		)

		jump := diff(pid, gid, pid1, gid1)
		if jump < T0x80 {
			jump -= 1
			h[0] = 0x80 | byte(jump) // T.2
			copy(h[1:], p)
			p = p[7:]
		} else if jump < T0xC0 {
			jump -= T0x80
			h[0] = 0xC0 | byte(jump>>8) // T.3.0
			h[1] = byte(jump)
			copy(h[2:], p)
			p = p[6:]
		} else if jump < T0xD0 {
			jump -= T0xC0
			h[0] = 0xD0 | byte(jump>>16) // T.3.10
			h[1] = byte(jump >> 8)
			h[2] = byte(jump)
			copy(h[3:], p)
			p = p[5:]
		} else if jump < T0xD8 {
			jump -= T0xD0
			h[0] = 0xD8 | byte(jump>>24) // T.3.11
			h[1] = byte(jump >> 16)
			h[2] = byte(jump >> 8)
			h[3] = byte(jump)
			copy(h[4:], p)
			p = p[4:]
		} else {
			jump = pack(pid1, gid1)
			h[0] = 0xE0 | byte(jump>>56) // T.4
			h[1] = byte(jump >> 48)
			h[2] = byte(jump >> 40)
			h[3] = byte(jump >> 32)
			h[4] = byte(jump >> 24)
			h[5] = byte(jump >> 16)
			h[6] = byte(jump >> 8)
			h[7] = byte(jump)
		}

		pid, gid = pid1, gid1
	}
}

func diff(pid uint64, gid uint16, pid1 uint64, gid1 uint16) uint64 {
	return pack(pid1, gid1) - pack(pid, gid)
}

func (a *Linked) need(size int, pid uint64, gid uint16) int {
	for i := 0; i < size; i++ {
		gid++
		if gid == pageGranules {
			pid++
			gid = 0
			if pid == uint64(len(a.pages)) {
				a.alloc()
			}
		}
		if a.occupied(pid, gid) {
			return i
		}
	}

	return size // TODO via bitset1
}

func (a *Linked) granule(pid uint64, gid uint16) *granule {
	return &a.pages[pid][gid]
}

func (a *Linked) occupied(pid uint64, gid uint16) bool {
	return a.bitset2[pid][(gid>>6)]&(1<<(gid&63)) != 0
}

func (a *Linked) occupy(pid uint64, gid uint16) {
	a.mark(pid, gid, true)
}

func (a *Linked) len() uint64 { return uint64(len(a.pages)) }
