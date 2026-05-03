package arena

import "math/bits"

type Sized struct {
	Linked
}

//| Класс  | гранул | Размер в байтах | Объектов на страницу |
//|--------|--------|-----------------|----------------------|
//| 0      | 1      | 8 B             | 16384                | -
//| 1      | 2      | 16 B            | 8192                 | -
//| 2      | 4      | 32 B            | 4096                 | 1 	3-4 	0,1
//| 3      | 8      | 64 B            | 2048                 | 1 	5-8 	0,1,2,3
//| 4      | 16     | 128 B           | 1024                 | 1 	9-16	0,1,2,3,4,5,6,7
//| 5      | 32     | 256 B           | 512                  |
//| 6      | 64     | 512 B           | 256                  |
//| 7      | 128    | 1 KB            | 128                  |
//| 8      | 256    | 2 KB            | 64                   |
//| 9      | 512    | 4 KB            | 32                   |
//| 10     | 1024   | 8 KB            | 16                   |
//| 11     | 2048   | 16 KB           | 8                    |
//| 12     | 4096   | 32 KB           | 4                    |
//| 13     | 8192   | 64 KB           | 2                    |
//| 14     | 16384  | 128 KB          | 1 (вся страница)     |
//
//| CID     | гранул | Шаг хвоста (гранул) | хвост                                         |
//|---------|--------|---------------------|-----------------------------------------------|
//| 0       | 1      | —                   | -                                             |
//| 1       | 2      | —                   | -                                             |
//| 2       | 4      | 1                   | 0, 1                                          |
//| 3       | 8      | 1                   | 0, 1, 2, 3                                    |
//| 4       | 16     | 1                   | 16, 15, 14... (шаг 1, используем 8 состояний) |
//| 5       | 32     | 2                   | 32, 30, 28, 26, 24, 22, 20, 18                |
//| 6       | 64     | 4                   | 64, 60, 56, 52, 48, 44, 40, 36                |
//| 7       | 128    | 8                   | 128, 120, 112, 104, 96, 88, 80, 72            |
//| ...     | ...    | ...                 | ...                                           |
//| 14      | 16384  | 1024                | 16384, 15360, ... 9216                        |

func class(n int) (cid uint8, sub uint8) {
	if n <= 1 {
		return 0, 0
	}

	// Находим базовый cid (степень двойки, куда влезает n)
	cid = uint8(bits.Len(uint(n - 1)))
	baseSize := 1 << cid

	// Определяем шаг подкласса (сколько гранул в одном sub)
	step := 1
	if cid > 3 {
		step = 1 << (cid - 3)
	}

	// Считаем, сколько гранул можно оставить свободными
	freeTail := baseSize - n
	sub = uint8(freeTail / step)

	// Ограничение в 3 бита
	if sub > 7 {
		sub = 7
	}

	return cid, sub
}

func (a *Sized) findExactly(n int) uint64 {
	cid, sub := class(n)
	size := (1 << cid) - (int(sub) * (1 << max(0, int(cid)-3)))

	align := uint16(1 << cid)
	pid, gid := unpack(a.hint)
	gid = (gid + align - 1) & ^(align - 1)

	for {
		if pid >= a.len() {
			pid, gid = a.alloc()
		} else {
			if gid+uint16(size) > pageGranules {
				pid++
				gid = 0
				continue
			}
			if !a.isAreaFree(pid, gid, size) {
				gid += align
				continue
			}
		}

		// Помечаем в битсете
		a.mark2(pid, gid, size)

		// Упаковываем VID
		// Собираем lowBits (15 бит с маркером)
		lowBits := ((gid >> cid) << (cid + 1)) | (1 << cid)

		// Распределяем 15 бит: 14 в хвост, 1 в 60-й бит
		vid := (uint64(sub) << 61) |
			(uint64(lowBits&0x4000) << (60 - 14)) |
			(pid << 14) |
			uint64(lowBits&0x3FFF)

		return vid
	}
}

func (a *Sized) isAreaFree(pid uint64, gid uint16, size interface{}) bool {
	return false
}

func (a *Sized) findExactly1(n int) (vid uint64) {
	cid, sub := class(n)

	// Реальный размер, который мы пометим в битсете
	step := 0
	if cid >= 3 {
		step = 1 << (cid - 3)
	}
	occupied := (1 << cid) - (int(sub) * step)

	// Параметры поиска
	align := uint16(1 << cid)
	pid, gid := unpack(a.hint)
	gid = (gid + align - 1) & ^(align - 1) // Выравниваем старт

	for {
		if pid >= a.len() {
			pid, gid = a.alloc() // Новая страница всегда чистая и выровненная
		} else {
			// Проверяем, влезает ли блок в страницу
			if gid+align > pageGranules {
				pid++
				gid = 0
				continue
			}
			// Проверяем битсет для нужного количества гранул
			if !a.isAreaFree(pid, gid, occupied) {
				gid += align
				continue
			}
		}

		// Если нашли — помечаем!
		a.mark2(pid, gid, occupied)

		// Обновляем хинт только если мы заняли самую раннюю позицию
		if pack(pid, gid) == a.hint {
			// В идеале тут нужно найти следующую дырку, но пока просто сдвинем
			a.hint = pack(pid, gid+uint16(occupied))
		}

		// УПАКОВКА VID (64 бита)
		// 63: 1 (Sized)
		// 62-60: sub
		// 59-14: pid
		// 13-0: packed gid (с магическим битом cid)

		packedGid := (gid << 1) | (1 << cid)
		if cid == 0 {
			// Для cid=0 packedGid будет (gid << 1) | 1.
			// TrailingZeros даст 0, (raw &^ 1) >> 1 вернет исходный gid.
		}

		vid = (1 << 63) | (uint64(sub) << 60) | (pid << 14) | uint64(packedGid>>1)
		// Уточнение: твой gid 14 бит, мы сдвигаем его в 14-битное поле.
		// Так как (gid << 1) | (1 << cid) может занять 15 бит (при cid=14),
		// нам нужно аккуратно вписать это в 14-15 бит.

		return vid
	}
}
