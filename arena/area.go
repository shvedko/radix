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

type Linked struct {
	bitset0 []uint64
	bitset1 []*bitset256
	bitset2 []*bitset16k
	pages   []*page
	pid     uint64
	gid     uint16
}

func (a *Linked) pack(pid uint64, gid uint16) uint64 { return (pid << 14) | uint64(gid) }
func (a *Linked) unpack(id uint64) (uint64, uint16)  { return id >> 14, uint16(id & 0x3FFF) }

func (a *Linked) Write(data []byte) uint64 {
	pid, gid := a.findFree(a.pid, a.gid)
	rid := a.pack(pid, gid)

	for {
		a.mark(pid, gid, true)
		curr := &a.pages[pid][gid]

		// 1. Конец цепочки
		if len(data) <= 7 {
			curr[0] = byte(len(data)) // Type 00 (биты 7-6) + 6 бит длины
			copy(curr[1:], data)
			break
		}

		// Ищем место для следующего куска
		nextPid, nextGIdx := a.findFree(pid, gid+1)

		// 2. Близко в той же странице (Type 01: 1 байт заголовка)
		diff := nextGIdx - gid
		if nextPid == pid && diff > 0 && diff < 64 {
			curr[0] = 0x40 | byte(diff) // 01000000 | delta
			n := copy(curr[1:], data[:7])
			data = data[n:]
			pid, gid = nextPid, nextGIdx
			continue
		}

		// 3. Далеко в той же странице (Type 10: 2 байта заголовка)
		if nextPid == pid {
			curr[0] = 0x80 | byte(nextGIdx>>8) // 10000000 | high 6 bits
			curr[1] = byte(nextGIdx)           // low 8 bits
			n := copy(curr[2:], data[:6])
			data = data[n:]
			pid, gid = nextPid, nextGIdx
			continue
		}

		// 4. Другая страница (Type 11: Jump-гранула)
		// Пишем 11 в заголовок текущей, но данных не кладем,
		// а следующей выбираем ту же самую, но в новой странице
		curr[0] = 0xC0
		nextAddr := a.pack(nextPid, nextGIdx)
		// Упаковываем 64-битный адрес в оставшиеся 7 байт + 6 бит заголовка
		// Для простоты: пишем адрес в [1:8]
		*(*uint64)(unsafe.Pointer(&curr[0])) |= (nextAddr << 8) // Условно

		pid, gid = nextPid, nextGIdx
		// В этой итерации data не уменьшилась, мы просто "прыгнули"
	}
	return rid
}

func (a *Linked) findFree(startPage uint64, startIdx uint16) (uint64, uint16) {
	// 1. Ищем строго от хинта и только вперед
	pid, gid, ok := a.scanForward(startPage, startIdx)
	if ok {
		a.pid, a.gid = pid, gid
		return pid, gid
	}

	// 2. Если не нашли — значит, реально всё занято до самого конца.
	// Добавляем новую страницу.
	pid, gid = a.alloc()
	a.pid, a.gid = pid, gid
	return pid, gid
}

// scanForward — выделенная логика сканирования bitset2 и bitset0
func (a *Linked) scanForward(pid uint64, gid uint16) (uint64, uint16, bool) {
	// Логика поиска по bitset в текущей странице...
	if p, i, ok := a.searchInPage(pid, gid); ok {
		return p, i, true
	}

	// Логика поиска через bitset0...
	startWord := (pid + 1) >> 6
	for w := startWord; w < uint64(len(a.bitset0)); w++ {
		mask := a.bitset0[w]
		if mask != ^uint64(0) {
			nextPage := (w << 6) + uint64(bits.TrailingZeros64(^mask))
			if p, i, ok := a.searchInPage(nextPage, 0); ok {
				return p, i, true
			}
		}
	}
	return 0, 0, false
}

//func (a *Linked) searchInPage(pid uint64, startIdx int) (uint64, int, bool) {
//	if pid >= uint64(len(a.bitset2)) {
//		return 0, 0, false
//	}
//	bs := &a.bitset2[pid]
//
//	startWord := startIdx >> 6
//	for i := startWord; i < 256; i++ {
//		mask := bs[i]
//		if i == startWord {
//			// Игнорируем биты МЕНЬШЕ startIdx
//			// Например, если startIdx = 3, маска станет 111...1000 (биты 0,1,2 заняты для нас)
//			mask |= (1 << (startIdx & 63)) - 1
//		}
//
//		if mask != ^uint64(0) {
//			idx := (i << 6) + bits.TrailingZeros64(^mask)
//			return pid, idx, true
//		}
//	}
//	return 0, 0, false
//}

func (a *Linked) searchInPage(pid uint64, gid uint16) (uint64, uint16, bool) {
	if pid >= uint64(len(a.bitset2)) {
		return 0, 0, false
	}
	bitset := a.bitset2[pid]
	for i := gid >> 6; i < 256; i++ {
		if bitset[i] != ^uint64(0) {
			idx := i<<6 + uint16(bits.TrailingZeros64(^bitset[i]))
			return pid, idx, true
		}
	}
	return 0, 0, false
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
