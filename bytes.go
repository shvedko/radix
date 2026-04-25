package radix

import (
	"encoding/binary"
	"math"
	"math/bits"
	"time"
	"unsafe"
)

func Float64ToKey(f float64) []byte {
	// 1. Берем битовое представление float как uint64
	u := math.Float64bits(f)

	// 2. Трансформируем биты
	if u>>63 == 0 {
		// Если число положительное (или +0):
		// Просто инвертируем знаковый бит (делаем его 1)
		u ^= 1 << 63
	} else {
		// Если число отрицательное:
		// Инвертируем ВООБЩЕ ВСЕ биты
		u = ^u
	}

	// 3. Переворачиваем в Big-Endian
	be := bits.ReverseBytes64(u)
	return unsafe.Slice((*byte)(unsafe.Pointer(&be)), 8)
}

func KeyToFloat64(b []byte) float64 {
	u := bits.ReverseBytes64(binary.BigEndian.Uint64(b))

	if u>>63 == 1 {
		// Было положительным — возвращаем знаковый бит на 0
		u ^= 1 << 63
	} else {
		// Было отрицательным — инвертируем всё обратно
		u = ^u
	}
	return math.Float64frombits(u)
}

func IntToKey(i int) []byte {
	// 1. Приводим к uint64 (размер int на 64-бит системах)
	b := uint64(i)

	// 2. Инвертируем знаковый бит (XOR с 1000...000)
	// Это превратит самое минимальное отрицательное число в 0,
	// а 0 — в середину диапазона.
	b ^= 1 << 63

	// 3. Переворачиваем в Big-Endian для правильной сортировки
	be := bits.ReverseBytes64(b)

	// 4. Возвращаем слайс от стековой переменной
	return unsafe.Slice((*byte)(unsafe.Pointer(&be)), 8)
}

func TimeToKey(t time.Time) []byte {
	// 1. Переводим в UTC и берем наносекунды
	// Это дает нам абсолютную точку на шкале времени
	nano := uint64(t.UTC().UnixNano())

	// 2. Поскольку время всегда положительное (относительно 1970 года),
	// нам достаточно просто перевернуть байты в Big-Endian.
	be := bits.ReverseBytes64(nano)

	return unsafe.Slice((*byte)(unsafe.Pointer(&be)), 8)
}
