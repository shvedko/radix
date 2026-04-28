package arena

import (
	"math"
	"math/bits"
	"unsafe"
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

func (a *Linked) pack(pid uint64, gid uint16) uint64 { return (pid << 14) | uint64(gid) }
func (a *Linked) unpack(id uint64) (uint64, uint16)  { return id >> 14, uint16(id & 0x3FFF) }

func (a *Linked) Write(p []byte) uint64 {
	pid, gid := a.next(a.pid, a.gid)
	rid := a.pack(pid, gid)

	for {
		a.mark(pid, gid, true)
		g := &a.pages[pid][gid]

		// 1. Конец цепочки (Type 00: (биты 7-6) + 6 бит длины)
		if len(p) <= 7 {
			g[0] = byte(len(p))
			copy(g[1:], p)
			break
		}

		// Ищем место для следующего куска
		pid1, gid1 := a.next(pid, gid+1)

		// 2. FIXME  Близко в той же странице (Type 01: 1 байт заголовка)
		diff := gid1 - gid
		if pid1 == pid && diff > 0 && diff < 64 {
			g[0] = 0x40 | byte(diff) // 01000000 | delta
			n := copy(g[1:], p[:7])
			p = p[n:]
			pid, gid = pid1, gid1
			continue
		}

		// 3. Далеко в той же странице (Type 10: 2 байта заголовка)
		if pid1 == pid {
			g[0] = 0x80 | byte(gid1>>8) // 10000000 | high 6 bits
			g[1] = byte(gid1)           // low 8 bits
			n := copy(g[2:], p[:6])
			p = p[n:]
			pid, gid = pid1, gid1
			continue
		}

		// 4. Другая страница (Type 11: Jump-гранула)
		// Пишем 11 в заголовок текущей, но данных не кладем,
		// а следующей выбираем ту же самую, но в новой странице
		g[0] = 0xC0
		nextRID := a.pack(pid1, gid1)
		// Упаковываем 64-битный адрес в оставшиеся 7 байт + 6 бит заголовка
		// Для простоты: пишем адрес в [1:8]
		*(*uint64)(unsafe.Pointer(&g[0])) |= nextRID << 8 // Условно

		pid, gid = pid1, gid1
		// В этой итерации p не уменьшилась, мы просто "прыгнули"
	}

	a.pid, a.gid = pid, gid

	return rid
}

func (a *Linked) next(pid uint64, gid uint16) (uint64, uint16) {
	pid, gid, ok := a.scan(pid, gid)
	if ok {
		return pid, gid
	}

	return a.alloc()
}

func (a *Linked) scan(pid uint64, gid uint16) (uint64, uint16, bool) {
	// 1. Ищем гранулу по bitset1 bitset2 в текущей странице...
	gid, ok := a.find(pid, gid)
	if ok {
		return pid, gid, true
	}

	// 2. Если не нашли, листаем страницы через bitset0...
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
			mask1 |= (1 << (gid2 & 63)) - 1
		}
		if mask1 == ^uint64(0) {
			continue
		}
		idx2 := i<<6 + uint16(bits.TrailingZeros64(^mask1))
		mask2 := a.bitset2[pid][idx2]
		if idx2 == gid2 {
			mask2 |= (1 << (gid & 63)) - 1
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
	pid := uint64(len(a.pages))
	if pid%64 == 0 {
		a.bitset0 = append(a.bitset0, 0)
	}
	a.bitset1 = append(a.bitset1, new(bitset256))
	a.bitset2 = append(a.bitset2, new(bitset16k))
	a.pages = append(a.pages, new(page))

	return pid, 0
}

func (a *Linked) Free(id uint64) {
	pid, gid := a.unpack(id)

	// 1. Обновляем хинт аллокатора только один раз.
	// Самая первая гранула в цепочке — всегда потенциально самая левая.
	if pid < a.pid || (pid == a.pid && gid < a.gid) {
		a.pid = pid
		a.gid = gid
	}

	// 2. Спокойно чистим цепочку
	for {
		curr := &a.pages[pid][gid]
		mode := curr[0] >> 6

		var pid1 uint64
		var gid1 uint16

		switch mode {
		case 0: // End
			a.mark(pid, gid, false)
			return
		case 1: // Near
			pid1 = pid
			gid1 = gid + uint16(curr[0]&0x3F)
		case 2: // Intra-page
			pid1 = pid
			gid1 = (uint16(curr[0]&0x3F) << 8) | uint16(curr[1])
		case 3: // Jump
			id = *(*uint64)(unsafe.Pointer(curr)) >> 8
			pid1, gid1 = a.unpack(id)
		}

		a.mark(pid, gid, false)

		pid, gid = pid1, gid1
	}
}

type cursor struct {
	a   *Linked
	pid uint64
	gid uint16
	off uint8 // Остаток данных в текущей грануле (после заголовка)
}

func (a *Linked) open(id uint64) cursor {
	pid, gid := a.unpack(id)
	return cursor{a: a, pid: pid, gid: gid}
}

func (c *cursor) Read(p []byte) int {
	var n int
	for n < len(p) {
		curr := &c.a.pages[c.pid][c.gid]
		mode := curr[0] >> 6
		data := curr[0] & 0x3F // Для mode 0 это длина полезных данных

		var payload []byte
		var nextPID uint64
		var nextGIdx int

		switch mode {
		case 0: // End
			payload = curr[1 : 1+data]
			// Если мы уже всё вычитали из этой последней гранулы
			if int(c.off) >= len(payload) {
				return n
			}
		case 1: // Near
			payload = curr[1:]
			nextPID = c.pid
			nextGIdx = int(c.gid) + int(data)
		case 2: // Intra-page
			payload = curr[2:]
			nextPID = c.pid
			nextGIdx = (int(data) << 8) | int(curr[1])
		case 3: // Jump
			addr := *(*uint64)(unsafe.Pointer(curr)) >> 8
			c.pid, c.gid = c.a.unpack(addr)
			continue
		}

		src := payload[c.off:]
		copied := copy(p[n:], src)
		n += copied
		c.off += uint8(copied)

		if int(c.off) < len(payload) {
			// p заполнился раньше, чем кончилась гранула
			return n
		}

		// Если это была последняя гранула, выходим (c.off останется равным len)
		if mode == 0 {
			return n
		}

		// Переходим к следующей грануле
		c.off, c.pid, c.gid = 0, nextPID, uint16(nextGIdx)
	}
	return n
}

func NewLinked(pages int) *Linked {
	var a Linked
	for i := 0; i < pages; i++ {
		a.alloc()
	}
	return &a
}
