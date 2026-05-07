package arena

import (
	"math/bits"
	"unsafe"
)

const (
	granuleBytes = 8
	pageGranules = 16384
	pageBytes    = pageGranules * granuleBytes
	_            = pageBytes / 1024
)

type granule [granuleBytes]byte
type page [pageGranules]granule
type bitset16k [pageGranules / 64]uint64
type bitset256 [pageGranules / 64 / 64]uint64

type Linked struct {
	bitset0 []uint64
	bitset1 []*bitset256
	bitset2 []*bitset16k
	pages   []*page
	hint    uint64
}

func pack(pid uint64, gid uint16) uint64 { return (pid << 14) | uint64(gid) }
func unpack(id uint64) (uint64, uint16)  { return id >> 14, uint16(id & 0x3FFF) }

func diff(pid uint64, gid uint16, pid1 uint64, gid1 uint16) uint64 {
	return pack(pid1, gid1) - pack(pid, gid)
}

func add(pid uint64, gid uint16, add uint64) (uint64, uint16) {
	return unpack(pack(pid, gid) + add)
}

func (a *Linked) next(pid uint64, gid uint16) (uint64, uint16) {
	if pid >= a.len() {
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
			if pid == a.len() {
				break
			}
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

func (a *Linked) mark2(pid uint64, gid uint16, granules uint16) {
	gid1 := gid + granules - 1
	idx1 := gid >> 6
	idx2 := gid1 >> 6

	for i := idx1; i <= idx2; i++ {
		bit1 := uint16(0)
		if i == idx1 {
			bit1 = gid & 63
		}

		bit2 := uint16(63)
		if i == idx2 {
			bit2 = gid1 & 63
		}

		j := i & 0xFF
		k := pid + uint64(i>>8)

		a.bitset2[k][j] |= (^uint64(0) >> (63 - (bit2 - bit1))) << bit1
		if a.bitset2[k][j] == ^uint64(0) {
			idx := j >> 6
			a.bitset1[k][idx] |= uint64(1) << (j & 63)
			if a.bitset1[k][idx] == ^uint64(0) {
				if a.bitset1[k][0] == ^uint64(0) &&
					a.bitset1[k][1] == ^uint64(0) &&
					a.bitset1[k][2] == ^uint64(0) &&
					a.bitset1[k][3] == ^uint64(0) {
					a.bitset0[k>>6] |= 1 << (k & 63)
				}
			}
		}
	}
}

func (a *Linked) unmark2(pid uint64, gid uint16, granules uint16) {
	gid1 := gid + granules - 1
	idx1 := gid >> 6
	idx2 := gid1 >> 6

	for i := idx1; i <= idx2; i++ {
		bit1 := uint16(0)
		if i == idx1 {
			bit1 = gid & 63
		}

		bit2 := uint16(63)
		if i == idx2 {
			bit2 = gid1 & 63
		}

		j := i & 0xFF
		k := pid + uint64(i>>8)

		a.bitset2[k][j] &^= (^uint64(0) >> (63 - (bit2 - bit1))) << bit1
		a.bitset1[k][(j >> 6)] &^= uint64(1) << (j & 63)
		a.bitset0[k>>6] &^= 1 << (k & 63)
	}
}

func (a *Linked) alloc() (uint64, uint16) {
	i, pid := 1, a.len()
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
		if a.len()&63 == 0 {
			a.bitset0 = append(a.bitset0, 0)
		}
		a.bitset1 = append(a.bitset1, &bits1[i])
		a.bitset2 = append(a.bitset2, &bits2[i])
		a.pages = append(a.pages, &pages[i])
	}

	return pid, 0
}

func (a *Linked) write(p []byte) uint64 {
	pid, gid := a.next(unpack(a.hint))
	rid := pack(pid, gid)

	for {
		a.mark(pid, gid, true)
		h := a.granule(pid, gid)

		if len(p) < 8 {
			h[0] = 0xF0 | byte(len(p)) // T.5
			copy(h[1:], p)
			a.hint = 1 + pack(pid, gid)
			return rid
		}

		size := uint16(min(129, len(p)/8))
		if size > 1 {
			size = a.need(size, pid, gid)
			if size > 1 {
				h[0] = byte(size - 2) // T.1
				copy(h[1:], p)
				p = p[7:]

				i := uint16(1)
				for i < size {
					gid++
					if gid == pageGranules {
						pid++
						gid = 0
					}

					x := min(size-i, pageGranules-gid)
					a.mark2(pid, gid, x)

					b := unsafe.Slice(&a.pages[pid][gid][0], x<<3)
					copy(b, p)
					p = p[x<<3:]

					i += x
					gid += x - 1
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
			T0x80 = 0x3F + 1 + 1          // 1000001
			T0xC0 = 0xFFF + T0x80 + 1     // 1000001000001
			T0xD0 = 0x7FFFF + T0xC0 + 1   // 10000001000001000001
			T0xD8 = 0x7FFFFFF + T0xD0 + 1 // 1000000010000001000001000001
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

func (a *Linked) need(size uint16, pid uint64, gid uint16) uint16 {
	var i uint16
	for {
		gid++
		if gid >= pageGranules {
			pid++
			gid = 0
			if pid >= a.len() {
				a.alloc()
			}
		}

		mask := a.bitset2[pid][(gid>>6)] >> (gid & 63)
		if mask == ^uint64(0) {
			return i
		}

		run := uint16(bits.TrailingZeros64(mask))
		end := 64 - (gid & 63)
		if run > end {
			run = end
		}

		i += run
		if i >= size {
			return size
		}

		if run < end {
			return i
		}

		gid += run - 1
	}
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

func (a *Linked) reset() {
	for i := range a.bitset1 {
		a.bitset1[i][0] = 0
		a.bitset1[i][1] = 0
		a.bitset1[i][2] = 0
		a.bitset1[i][3] = 0
		for j := range a.bitset2[i] {
			a.bitset2[i][j] = 0
		}
	}
	for i := range a.bitset0 {
		a.bitset0[i] = 0
	}
	a.hint = 0
}

type cursor struct {
	a   *Linked
	pid uint64
	gid uint16
	rem uint16
	off uint8
}

func (a *Linked) open(id uint64) cursor {
	pid, gid := unpack(id)
	return cursor{
		a:   a,
		pid: pid,
		gid: gid,
	}
}

func (c *cursor) read(p []byte) int {
	var n int
	for n < len(p) {
		h := c.a.granule(c.pid, c.gid)

		var jump uint64
		if c.rem > 0 {
			m := min(c.rem, pageGranules-c.gid)
			b := unsafe.Slice(&h[0], m<<3)
			x := copy(p[n:], b[c.off:])
			n += x
			x += int(c.off)
			c.off = uint8(x & 7)
			x >>= 3
			c.rem -= uint16(x)
			if c.off > 0 {
				break
			}
			jump = uint64(x)
		} else if h[0]&0xF8 == 0xF0 { // T.5 [11110...]
			n += c.copy(p[n:], h[1+c.off:1+h[0]&0x07])
			break
		} else if h[0]&0x80 == 0x00 { // T.1 [0.......]
			n += c.copy(p[n:], h[1+c.off:])
			if c.off < 7 {
				break
			} else {
				c.rem = 1 + uint16(h[0])
			}
			jump = 1
		} else if h[0]&0xC0 == 0x80 { // T.2 [10......]
			n += c.copy(p[n:], h[1+c.off:])
			if c.off < 7 {
				break
			}
			jump = uint64(h[0] & 0x3F)
			jump += 0b1
		} else if h[0]&0xF0 == 0xC0 { // T.3.0 [1100....][........]
			n += c.copy(p[n:], h[2+c.off:])
			if c.off < 6 {
				break
			}
			jump = uint64(h[0]&0x0F)<<8 | uint64(h[1])
			jump += 0b1000001
		} else if h[0]&0xF8 == 0xD0 { // T.3.10 [11010...][........][........]
			n += c.copy(p[n:], h[3+c.off:])
			if c.off < 5 {
				break
			}
			jump = uint64(h[0]&0x07)<<16 | uint64(h[1])<<8 | uint64(h[2])
			jump += 0b1000001000001
		} else if h[0]&0xF8 == 0xD8 { // T.3.11 [11011...][........][........][........]
			n += c.copy(p[n:], h[4+c.off:])
			if c.off < 4 {
				break
			}
			jump = uint64(h[0]&0x07)<<24 | uint64(h[1])<<16 | uint64(h[2])<<8 | uint64(h[3])
			jump += 0b10000001000001000001
		} else if h[0]&0xF0 == 0xE0 { // T.4 [1110....][7]
			id := uint64(h[0]&0x0F)<<56 |
				uint64(h[1])<<48 | uint64(h[2])<<40 |
				uint64(h[3])<<32 | uint64(h[4])<<24 |
				uint64(h[5])<<16 | uint64(h[6])<<8 | uint64(h[7])
			c.pid, c.gid = unpack(id)
			c.off = 0
			continue
		} else {
			return -1
		}

		c.off = 0
		c.pid, c.gid = add(c.pid, c.gid, jump)
	}

	return n
}

func (c *cursor) copy(p, b []byte) int {
	n := copy(p, b)
	c.off += uint8(n)
	return n
}

func (a *Linked) free(id uint64) {
	pid, gid := unpack(id)

	var rem uint16
	for {
		h := a.granule(pid, gid)

		var jump uint64
		if rem > 0 {
			x := min(rem, pageGranules-gid)
			rem -= x
			a.unmark2(pid, gid, x)
			jump = uint64(x)
		} else if h[0]&0xF8 == 0xF0 { // T.5 [11110...]
			a.mark(pid, gid, false)
			if a.hint >= id {
				a.hint = id
			}
			break
		} else if h[0]&0x80 == 0x00 { // T.1 [0.......]
			a.mark(pid, gid, false)
			rem = 1 + uint16(h[0])
			jump = 1
		} else if h[0]&0xC0 == 0x80 { // T.2 [10......]
			a.mark(pid, gid, false)
			jump = uint64(h[0] & 0x3F)
			jump += 0b1
		} else if h[0]&0xF0 == 0xC0 { // T.3.0 [1100....][........]
			a.mark(pid, gid, false)
			jump = uint64(h[0]&0x0F)<<8 | uint64(h[1])
			jump += 0b1000001
		} else if h[0]&0xF8 == 0xD0 { // T.3.10 [11010...][........][........]
			a.mark(pid, gid, false)
			jump = uint64(h[0]&0x07)<<16 | uint64(h[1])<<8 | uint64(h[2])
			jump += 0b1000001000001
		} else if h[0]&0xF8 == 0xD8 { // T.3.11 [11011...][........][........][........]
			a.mark(pid, gid, false)
			jump = uint64(h[0]&0x07)<<24 | uint64(h[1])<<16 | uint64(h[2])<<8 | uint64(h[3])
			jump += 0b10000001000001000001
		} else if h[0]&0xF0 == 0xE0 { // T.4 [1110....][7]
			a.mark(pid, gid, false)
			pid, gid = unpack(
				uint64(h[0]&0x0F)<<56 |
					uint64(h[1])<<48 | uint64(h[2])<<40 |
					uint64(h[3])<<32 | uint64(h[4])<<24 |
					uint64(h[5])<<16 | uint64(h[6])<<8 | uint64(h[7]))
			continue
		} else {
			panic(h)
		}

		pid, gid = add(pid, gid, jump)
	}
}

func (a *Linked) move(pid uint64, gid uint16) {
	a.hint = pack(pid, gid)
}

func (a *Linked) unmark(pid uint64, gid uint16) {
	a.mark(pid, gid, false)
}
