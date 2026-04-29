package arena

func NewLinked(pages int) *Linked {
	var a Linked
	for i := 0; i < pages; i++ {
		a.alloc()
	}
	return &a
}

/*

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

//
//func (c *cursor) step() {
//    c.gid++
//    if c.gid == pageGranules {
//        c.pid++
//        c.gid = 0
//    }
//}
