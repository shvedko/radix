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

func (a *Linked) Write(p []byte) uint64 {
	pid, gid := a.next(a.pid, a.gid)
	rid := pack(pid, gid)

	for len(p) > 0 {
		a.mark(pid, gid, true)
		h := &a.pages[pid][gid]

		// 1. ПРОВЕРКА НА КОНЕЦ
		if len(p) <= 7 {
			h[0] = 0xf0 | byte(len(p)) // 11110 + 3 бита длины
			copy(h[1:], p)
			return rid
		}

		// 2. ПРОВЕРКА НА СТРИМ (Тип 0)
		// Ищем сколько гранул впереди свободны физически подряд
		size, pid1, gid1 := 1, pid, gid
		// Максимум 128 гранул в потоке (1 заголовочная + 127 безадресных)
		for size < 128 && len(p)-7 > (size*8-0) {
			// Вычисляем, где ДОЛЖНА быть следующая гранула физически
			pid2, gid2 := pid1, gid1+1
			if gid2 == pageGranules {
				pid2++
				gid2 = 0
			}

			// Проверяем, свободна ли она на самом деле
			nP, nG := a.next(pid1, gid1+1)

			// Если следующая свободная гранула — это именно та, что идет следом в памяти
			if nP == pid2 && nG == gid2 {
				size++
				pid1, gid1 = nP, nG
			} else {
				break // Разрыв: либо занято, либо прыжок через дырку
			}
		}

		if size > 1 {
			h[0] = byte(size - 1) // Заголовок стрима 0 + 7 бит (0..127)
			n := copy(h[1:], p[:7])
			p = p[n:]

			for i := 1; i < size; i++ {
				// Просто переходим к физически следующей грануле
				gid++
				if gid == pageGranules { // Переход границы страницы
					pid++
					gid = 0
				}

				a.mark(pid, gid, true) // Помечаем как занятую
				n = copy(a.pages[pid][gid][:], p)
				p = p[n:]
			}
			// После цикла нам нужно найти СЛЕДУЮЩУЮ свободную точку для следующей итерации
			pid, gid = a.next(pid, gid+1)
			continue
		}

		// 3. ПРЫЖКИ (Если стрим не получился)
		nextP, nextG := a.next(pid, gid+1)

		if nextP == pid {
			diff := nextG - gid // всегда >= 1

			if diff <= 64 {
				// Тип 2: Короткий (10 + 6 бит)
				h[0] = 0x80 | byte(diff-1)
				n := copy(h[1:], p[:7])
				p = p[n:]
			} else if diff <= 4160 {
				// Тип 3.0: 2 байта (110 + 0 + 12 бит)
				val := uint16(diff - 65)
				h[0] = 0xc0 | byte(val>>8)
				h[1] = byte(val)
				n := copy(h[2:], p[:6])
				p = p[n:]
			} else {
				// Тип 3.10 / 3.11 (3-4 байта) — аналогично с bias
				// ... (для краткости пропустим, логика та же)
			}
		} else {
			// Тип 4: Jump (1110)
			h[0] = 0xe0
			nextID := pack(nextP, nextG)
			*(*uint64)(unsafe.Pointer(&h[0])) |= nextID << 4 // 4 бита флаг
			// p не уменьшаем, просто прыгнули
		}
		pid, gid = nextP, nextG
	}

	return rid
}

func (a *Linked) granule(pid uint64, gid uint16) *granule { return &a.pages[pid][gid] }

func (a *Linked) bit2(pid uint64, gid uint16) bool {
	return a.bitset2[pid][(gid>>6)]&(1<<(gid&63)) != 0
}
