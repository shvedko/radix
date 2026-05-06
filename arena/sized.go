package arena

import (
	"math/bits"
)

type Sized struct {
	Linked
	hints [16]uint64
}

func get28(g *granule) (uint32, uint32) {
	if g[0]&0x80 == 0x00 { // 0.......
		return uint32(g[0]), 1
	}
	if g[0]&0xC0 == 0x80 { // 10...... ........
		return uint32(g[0]&0x3F)<<8 | uint32(g[1]), 2
	}
	if g[0]&0xE0 == 0xC0 { // 110..... ........ ........
		return uint32(g[0]&0x1F)<<16 | uint32(g[1])<<8 | uint32(g[2]), 3
	}
	if g[0]&0xF0 == 0xE0 { // 1110.... ........ ........ ........
		return uint32(g[0]&0x0F)<<24 | uint32(g[1])<<16 | uint32(g[2])<<8 | uint32(g[3]), 4
	}
	return 0, 0
}

func put28(g *granule, size uint32) uint32 {
	n := bits.Len32(size)
	if n < 8 {
		g[0] = byte(size)
		return 1
	}
	if n < 15 {
		g[0] = 0x80 | byte(size>>8)
		g[1] = byte(size)
		return 2
	}
	if n < 22 {
		g[0] = 0xC0 | byte(size>>16)
		g[1] = byte(size >> 8)
		g[2] = byte(size)
		return 3
	}
	if n < 29 {
		g[0] = 0xE0 | byte(size>>24)
		g[1] = byte(size >> 16)
		g[2] = byte(size >> 8)
		g[3] = byte(size)
		return 4
	}
	return 0
}

// class14
//
//	| class | granule | size   | piece  | range      | step | remain                               |
//	|-------|---------|--------|--------|------------|------|--------------------------------------|
//	| 0     | 1       | 8 B    | 16384  | 1-1        | 1    | 0                                    |
//	| 1     | 2       | 16 B   | 8192   | 2-2        | 1    | 0                                    |
//	| 2     | 4       | 32 B   | 4096   | 3-4        | 1    | 0,1                                  |
//	| 3     | 8       | 64 B   | 2048   | 5-8        | 1    | 0,1,2,3                              |
//	| 4     | 16      | 128 B  | 1024   | 9-16       | 1    | 0,1,2,3,4,5,6,7                      |
//	| 5     | 32      | 256 B  | 512    | 18-32      | 2    | 0,2,4,6,8,10,12,14                   |
//	| 6     | 64      | 512 B  | 256    | 36-64      | 4    | 0,4,8,12,16,20,24,28                 |
//	| 7     | 128     | 1 KB   | 128    | 72-128     | 8    | 0,8,16,24,32,40,48,56                |
//	| 8     | 256     | 2 KB   | 64     | 144-256    | 16   | 0,16,32,48,64,80,96,112              |
//	| 9     | 512     | 4 KB   | 32     | 288-512    | 32   | 0,32,64,96,128,160,192,224           |
//	| 10    | 1024    | 8 KB   | 16     | 576-1024   | 64   | 0,64,128,192,256,320,384,448         |
//	| 11    | 2048    | 16 KB  | 8      | 1152-2048  | 128  | 0,128,256,384,512,640,768,896        |
//	| 12    | 4096    | 32 KB  | 4      | 2304-4096  | 256  | 0,256,512,768,1024,1280,1536,1792    |
//	| 13    | 8192    | 64 KB  | 2      | 4608-8192  | 512  | 0,512,1024,1536,2048,2560,3072,3584  |
//	| 14    | 16384   | 128 KB | 1      | 9216-16384 | 1024 | 0,1024,2048,3072,4096,5120,6144,7168 |
//	|-------|---------|--------|--------|------------|------|--------------------------------------|
func class14(granules uint32) (class, remain, step uint32) {
	class = uint32(bits.Len32(granules - 1))
	if class < 2 {
		step = 1
	} else if class < 5 {
		step = 1
		remain = (1 << class) - (granules)
	} else if class < 15 {
		step = 1 << (class - 4)
		remain = ((1 << class) - (granules)) >> (class - 4)
	}
	return
}

// AllocAligned ищет или выделяет свободный блок размером size гранул,
// размер size обязан быть степенью двойки (1, 2, 4, ..., 16384).
// Возвращает ID страницы (pid) и ID первой гранулы (gid).
func (a *Sized) AllocAligned(class uint16) (uint64, uint16) {
	size := (8) << class

	// Проходим по существующим страницам сверху вниз
	for pid := uint64(0); pid < uint64(len(a.pages)); pid++ {
		// Быстрая проверка: если бит в bitset0 установлен, в странице нет места вообще
		if (a.bitset0[pid>>6] & (1 << (pid & 63))) != 0 {
			continue
		}

		gid, ok := a.want(pid, size)
		if ok {
			//	a.mark2(pid, gid, size)
			return pid, gid
		}
	}

	// Если места нет ни в одной странице, выделяем новую
	pid, gid := a.alloc()
	//	a.mark2(pid, gid, size)
	return pid, gid
}

// aligns -- содержит маски, где 1 стоит только на позициях, кратных размеру блока.
// Например, для size=8 это биты 0, 8, 16, 24, 32, 40, 48, 56.
var aligns = [65]uint64{
	1:  0xFFFFFFFFFFFFFFFF, // каждое смещение валидно
	2:  0x5555555555555555, // 01010101...
	4:  0x1111111111111111, // 00010001...
	8:  0x0101010101010101, // 00000001...
	16: 0x0001000100010001,
	32: 0x0000000100000001,
	64: 0x0000000000000001,
}

// want выполняет поиск внутри конкретной страницы
func (a *Sized) want(pid uint64, size int) (uint16, bool) {
	words := size >> 6

	// Случай 1: Блок помещается внутри одного uint64 (64 гранулы и меньше)
	if words == 0 {
		for i := 0; i < len(a.bitset2[pid]); i++ {
			mask := a.bitset2[pid][i]
			if mask == ^uint64(0) {
				continue
			}

			// SWAR алгоритм: "размазываем" занятые биты вправо.
			// Если блок размера size имеет хотя бы один занятый бит (1),
			// после этого цикла стартовый бит этого блока (кратный size) станет равен 1.
			for s := 1; s < size; s <<= 1 {
				mask |= mask >> s
			}

			// 2. Оставляем только биты, соответствующие кратным позициям (0, size, 2*size...)
			// Пример для size=4: mask & 0x1111111111111111
			// Формула маски: (111...) / ((1 << size) - 1)
			// alignmentMask := uint64(0xFFFFFFFFFFFFFFFF) / ((1 << size) - 1)

			// Оставляем только те биты, которые соответствуют выровненным позициям (0, size, 2*size...)
			// и инвертируем, чтобы найти свободные (0 -> 1)
			free := (^mask) & aligns[size]

			if free != 0 {
				// TrailingZeros64 находит индекс первого подходящего блока без циклов
				shift := uint16(bits.TrailingZeros64(free))
				return uint16(i*64) + shift, true
			}
		}
		return 0, false
	}

	// Случай 2: Блок больше одного uint64 (128, 256... до 16384 гранул)
	// Вычисляем, сколько целых слов uint64 в bitset2 занимает такой блок

	if words == 256 {
		if a.bitset0[pid>>6]&(1<<(pid&63)) == 0 &&
			a.bitset1[pid][0] == 0 &&
			a.bitset1[pid][1] == 0 &&
			a.bitset1[pid][2] == 0 &&
			a.bitset1[pid][3] == 0 {
			return 0, true
		}
		return 0, false
	}

	// Шагаем сразу по индексам, кратным количеству слов (обеспечивает выравнивание)
	for i := 0; i < 256; i += words {

		//// Проверяем, свободен ли весь диапазон слов.
		//// Обычно для больших блоков size это всего 2, 4 или 8 слов.
		//allFree := true
		//for j := 0; j < words; j++ {
		//	if a.bitset2[pid][i+j] != 0 {
		//		allFree = false
		//		break
		//	}
		//}
		//
		//if allFree {
		//	return uint16(i * 64), true
		//}

		// Быстрая проверка диапазона в bitset2
		if a.bitset2[pid].empty(i, i+words) {
			return uint16(i * 64), true
		}
	}

	return 0, false
}

func (b *bitset16k) empty(from, to int) bool {
	for i := range b[from:to] {
		if b[from:to][i] != 0 {
			return false
		}
	}
	return true
}

// isRangeFree проверяет, что в bitset2 на странице pid слова [start..start+count) равны 0
func (a *Linked) isRangeFree(pid uint64, start, count int) bool {
	// Для маленьких count (2, 4, 8) цикл эффективнее всего
	// Для больших count компилятор Go применит SIMD оптимизации
	target := a.bitset2[pid][start : start+count]
	for i := range target {
		if target[i] != 0 {
			return false
		}
	}
	return true
}

func (a *Sized) reset() {
	a.Linked.reset()
	a.hints = [16]uint64{}
}
