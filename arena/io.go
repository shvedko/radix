package arena

/*

func NewLinked(pages int) *Linked {
	var a Linked
	for i := 0; i < pages; i++ {
		a.alloc()
	}
	return &a
}

func (a *Linked) Write1(p []byte) uint64 {
	pid, gid := a.next(a.pid, a.gid)
	rootID := a.pack(pid, gid)

	for len(p) > 0 {
		a.mark(pid, gid, true)
		g := &a.pages[pid][gid]

		// 1. ПРОВЕРКА НА КОНЕЦ
		if len(p) <= 7 {
			g[0] = 0xf0 | byte(len(p)) // 11110 + 3 бита длины
			copy(g[1:], p)
			return rootID
		}

		// 2. ПРОВЕРКА НА СТРИМ (Тип 0)
		// Ищем сколько гранул впереди свободны физически подряд
		run := 1
		tmpP, tmpG := pid, gid
		// Максимум 128 гранул в стриме (1 заголовочная + 127 безадресных)
		for run < 128 && len(p) > (run*8-1) {
			nP, nG := a.next(tmpP, tmpG+1)
			if nP != tmpP || nG != tmpG+1 { // разрыв страницы или занято
				break
			}
			run++
			tmpP, tmpG = nP, nG
		}

		if run > 1 {
			g[0] = byte(run - 1) // 0 + 7 бит (0..127)
			n := copy(g[1:], p[:7])
			p = p[n:]
			for i := 1; i < run; i++ {
				pid, gid = a.next(pid, gid+1)
				a.mark(pid, gid, true)
				n = copy(a.pages[pid][gid][:], p)
				p = p[n:]
			}
			pid, gid = a.next(pid, gid+1)
			continue
		}

		// 3. ПРЫЖКИ (Если стрим не получился)
		nextP, nextG := a.next(pid, gid+1)

		if nextP == pid {
			diff := nextG - gid // всегда >= 1

			if diff <= 64 {
				// Тип 2: Короткий (10 + 6 бит)
				g[0] = 0x80 | byte(diff-1)
				n := copy(g[1:], p[:7])
				p = p[n:]
			} else if diff <= 4160 {
				// Тип 3.0: 2 байта (110 + 0 + 12 бит)
				val := uint16(diff - 65)
				g[0] = 0xc0 | byte(val>>8)
				g[1] = byte(val)
				n := copy(g[2:], p[:6])
				p = p[n:]
			} else {
				// Тип 3.10 / 3.11 (3-4 байта) — аналогично с bias
				// ... (для краткости пропустим, логика та же)
			}
		} else {
			// Тип 4: Jump (1110)
			g[0] = 0xe0
			nextID := a.pack(nextP, nextG)
			*(*uint64)(unsafe.Pointer(&g[0])) |= nextID << 4 // 4 бита флаг
			// p не уменьшаем, просто прыгнули
		}
		pid, gid = nextP, nextG
	}
	return rootID
}

type cursor1 struct {
	a      *Linked
	pid    uint64
	gid    uint16
	off    uint8
	stream uint8 // сколько еще гранул в стриме читать без заголовков
}

func (c *cursor1) Read1(p []byte) int {
	var n int
	for n < len(p) {
		g := &c.a.pages[c.pid][c.gid]
		var payload []byte
		var nextP uint64
		var nextG uint16

		if c.stream > 0 {
			// МЫ ВНУТРИ СТРИМА
			payload = g[:]
			c.stream--
			nextP, nextG = c.pid, c.gid+1 // физически следующая
		} else {
			// ЧИТАЕМ ЗАГОЛОВОК
			h := g[0]
			if h < 0x80 { // Тип 1: Stream (0...)
				c.stream = h & 0x7f
				payload = g[1:]
				nextP, nextG = c.pid, c.gid+1
			} else if h < 0xc0 { // Тип 2: Short (10...)
				dist := (h & 0x3f) + 1
				payload = g[1:]
				nextP, nextG = c.pid, c.gid+uint16(dist)
			} else if h < 0xe0 { // Тип 3: Medium (110...)
				// 110 + submode
				sub := (h >> 4) & 0x3
				if sub == 0 { // 110 0 + 12 бит
					dist := (uint16(h&0xf)<<8 | uint16(g[1])) + 65
					payload = g[2:]
					nextP, nextG = c.pid, c.gid+dist
				}
				// ... остальные подтипы
			} else if h < 0xf0 { // Тип 4: Jump (1110...)
				fullID := *(*uint64)(unsafe.Pointer(&g[0])) >> 4
				c.pid, c.gid = c.a.unpack(fullID)
				continue
			} else { // Тип 5: End (11110...)
				length := h & 0x07
				payload = g[1 : 1+length]
				if int(c.off) >= len(payload) {
					return n
				}
				n += copy(p[n:], payload[c.off:])
				return n
			}
		}


		src := payload[c.off:]
		cp := copy(p[n:], src)
		n += cp
		c.off += uint8(cp)

		if int(c.off) < len(payload) {
			return n
		}

		c.off = 0
		c.pid, c.gid = nextP, nextG
	}
	return n
}

func (a *Linked) Free1(id uint64) {
	pid, gid := a.unpack(id)
	// ... хинт (как в прошлый раз) ...

	for {
		g := &a.pages[pid][gid]
		h := g[0]

		var nextP uint64
		var nextG uint16
		var stream uint8

		if h < 0x80 { // Stream
			stream = h & 0x7f
			nextP, nextG = pid, gid+1
		} else if h < 0xc0 { // Short
			nextP, nextG = pid, gid+uint16((h&0x3f)+1)
		} // ... и так далее для всех типов

		a.mark(pid, gid, false)

		// Если это был стрим, освобождаем N гранул подряд
		for stream > 0 {
			pid, gid = nextP, nextG
			a.mark(pid, gid, false)
			nextP, nextG = pid, gid+1
			stream--
		}

		if h >= 0xf0 {
			return
		} // End
		pid, gid = nextP, nextG
	}
}

*/

//if run > 1 {
//    g[0] = byte(run - 1) // Заголовок стрима
//    n := copy(g[1:], p[:7])
//    p = p[n:]
//
//    for i := 1; i < run; i++ {
//        // Просто переходим к физически следующей грануле
//        gid++
//        if gid == pageGranules { // Переход границы страницы
//            pid++
//            gid = 0
//        }
//
//        a.mark(pid, gid, true) // Помечаем как занятую
//        n = copy(a.pages[pid][gid][:], p)
//        p = p[n:]
//    }
//    // После цикла нам нужно найти СЛЕДУЮЩУЮ свободную точку для следующей итерации
//    pid, gid = a.next(pid, gid + 1)
//    continue
//}
//
//for run < 128 && len(p) > (run*8 - 1) {
//    // Вычисляем, где ДОЛЖНА быть следующая гранула физически
//    targetP, targetG := tmpP, tmpG + 1
//    if targetG == pageGranules {
//        targetP++
//        targetG = 0
//    }
//
//    // Проверяем, свободна ли она на самом деле
//    nP, nG := a.next(tmpP, tmpG + 1)
//
//    // Если следующая свободная гранула — это именно та, что идет следом в памяти
//    if nP == targetP && nG == targetG {
//        run++
//        tmpP, tmpG = nP, nG
//    } else {
//        break // Разрыв: либо занято, либо прыжок через дырку
//    }
//}
//
//func (c *cursor) step() {
//    c.gid++
//    if c.gid == pageGranules {
//        c.pid++
//        c.gid = 0
//    }
//}
