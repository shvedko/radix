------------------------------
## Radix
Высокопроизводительное Sparse Radix Tree на языке Go. Оптимизировано для работы с наносекундными задержками, минимальным потреблением памяти и поддержкой многослойных префиксных запросов.
## 🚀 Особенности

* Использование битмапов (uint256) и компактных срезов вместо плотных массивов. 
* Итератор спроектирован так, чтобы минимизировать аллокации в куче при поиске и обходе.
* Многослойность "вертикальных" связей (слоев) позволяет использовать дерево как многоколоночный индекс.
* Удаление через итератор со встроенным механизмом схлопывания (Merge) узлов для поддержания идеальной плотности дерева.
* Экстремальная скорость:
* Поиск: ~230 ns
    * Обход (Dump): ~450 ns (на 100 элементов)
    * Удаление: ~15 ns (чистая операция)

## 🛠 Установка

```shell
go get github.com/shvedko/radix
```

## 📖 Примеры использования

## Базовые операции 

```go
package main

import (
	"fmt"
	"radix"
)

func main() {
	t := radix.New[int]()

	// Вставка данных (ключ может состоять из нескольких слоев)
	t.Insert(1, false, []byte("Pavlov"), []byte("Ivan"))
	t.Insert(2, false, []byte("Pavlov"), []byte("Igor"))
	t.Insert(3, false, []byte("Petrov"), []byte("Ivan"))
	t.Insert(4, false, []byte("Pavlov"), []byte("Igor"))
	t.Insert(5, false, []byte("Pavlov"), []byte("Igor"), []byte("Vasilievich"))
	
	// Визуализация структуры дерева
	t.Dump(func(prefix []byte, level uint32, end bool, values []int) bool {
		fmt.Printf("%*s %s: %v\n", level*2, "", string(prefix), values)
		return true
	})
	
	//  : [] 
	//  P: [] 
	//    avlov: [] 
	//      : [] 
	//        I: [] 
	//          gor: [2 4] 
	//            : [] 
	//              Vasilievich: [5] 
	//          van: [1] 
	//    etrov: [] 
	//      : [] 
	//        Ivan: [3]
}
```

## Сложный поиск и фильтрация
Итератор поддерживает nil в качестве wildcard для любого слоя.

```go
	// Найти всех с фамилией на "P" и именем на "I"
	it := t.Search([]byte("P"), []byte("I"))
	for it.Next() {
		fmt.Println("Found:", it.Get())
	}

	// Найти всех с любым первым полем, но именем "Ivan"
	it = t.Search(nil, []byte("Ivan"))
	for it.Next() {
		fmt.Println("Found on second layer:", it.Get())
	}
	
	// Found: [2 4] 
	// Found: [5] 
	// Found: [1] 
	// Found: [3] 
	// Found on second layer: [1] 
	// Found on second layer: [3]
```

## Удаление через итератор
Позволяет безопасно удалять элементы во время обхода с автоматической оптимизацией (схлопыванием) дерева.

```go
	it = t.Search([]byte("Pavlov"))
	for it.Next() {
		got := it.Get()
		if got[0] == 1 { // Нашли Ивана
			it.Remove()
			break
		}
	}
	
	// Вставка с получением итератора на узле 
	it, ok := t.InsertPath(1, false, []byte("Pavlov"), []byte("Ivan"))
	if ok {
		it.Remove() // Rollback
	}
```

## 📊 Benchmarks (Intel i7-11700F)

```
cpu: 11th Gen Intel(R) Core(TM) i7-11700F @ 2.50GHz
BenchmarkRadix_100/Search/First-16      12043537                87.58 ns/op          128 B/op          1 allocs/op
BenchmarkRadix_100/Search/Point-16       7284444               169.4 ns/op           176 B/op          2 allocs/op
BenchmarkRadix_100/Search/Prefix-16      6393817               188.5 ns/op           176 B/op          2 allocs/op
BenchmarkRadix_100/Search/Scan-16        1872069               636.7 ns/op           176 B/op          2 allocs/op
BenchmarkRadix_100/Dump-16               2655260               451.4 ns/op             0 B/op          0 allocs/op
BenchmarkRadix_100/Walk-16               2595984               461.4 ns/op             0 B/op          0 allocs/op
BenchmarkRadix_100/Insert-Delete-16      7764678               152.0 ns/op           178 B/op          2 allocs/op
BenchmarkRadix_100/Insert-Only-16       25192300                46.94 ns/op            0 B/op          0 allocs/op
BenchmarkRadix_100/Insert-GoMap-16      17790534                62.67 ns/op           68 B/op          1 allocs/op
```

## ⚠️ Важно

1. Владение памятью: Дерево не копирует префиксы при вставке (Zero-copy стратегия). Убедитесь, что исходные данные не изменяются во время жизни дерева.
2. Потокобезопасность: Пакет **не является потокобезопасным**. Используйте sync.RWMutex или шардирование для конкурентного доступа.
