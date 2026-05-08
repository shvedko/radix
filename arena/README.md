Linked Arena — это zero-allocation аллокатор для Go, работающий на скоростях до 30 ГБ/с. Идеален для систем, где стандартный append в слайс и работа GC становятся узким местом.

-------------

Базовый цикл: Запись — Чтение — Удаление

```go
arena := &Linked{}
// 1. Быстрая запись (12 ГБ/с)
// Возвращает ID, который компактно упакован (pid + gid)
id := arena.Write([]byte("high performance data"))
// 2. Чтение через стековый курсор (30 ГБ/с)
// Курсор не аллоцирует память в куче
c := arena.Open(id)
buf := make([]byte, 1024)
n := c.Read(buf)
// 3. Мгновенное освобождение (70 ГБ/с)
// Место сразу становится доступным для новых записей
arena.Free(id)
```

Работа с фрагментированной памятью (Sparse Arena)

```go
// Занимаем место в середине страницы вручную
arena.Occupy(0, 100)
// Запись автоматически найдет дырку, создаст "прыжок" (T.2/T.3)
// и вернет единый ID для логически целых данных
id := arena.Write(largeData)
```

-------------

``` 
goos: windows
goarch: amd64
pkg: github.com/shvedko/radix/arena
cpu: 11th Gen Intel(R) Core(TM) i7-11700F @ 2.50GHz
BenchmarkLinked_find/emptied-16                 393237880                3.011 ns/op           0 B/op          0 allocs/op
BenchmarkLinked_find/occupied-16                319741221                3.823 ns/op           0 B/op          0 allocs/op
BenchmarkLinked_mark-16                         519429250                2.298 ns/op           0 B/op          0 allocs/op
BenchmarkLinked_write/copy-16                    1310485               923.1 ns/op      70992.31 MB/s          0 B/op          0 allocs/op
BenchmarkLinked_write/1KB-16                    13922990                82.24 ns/op     12450.76 MB/s         78 B/op          0 allocs/op
BenchmarkLinked_write/64KB-16                     220128              5150 ns/op        12725.42 MB/s       4964 B/op          0 allocs/op
BenchmarkLinked_write/512KB-16                     27499             43218 ns/op        12131.34 MB/s      39738 B/op          0 allocs/op
BenchmarkLinked_write/1MB-16                       13912             86796 ns/op        12080.91 MB/s      78702 B/op          0 allocs/op
BenchmarkLinked_write/8MB-16                        1624            717375 ns/op        11693.48 MB/s     678798 B/op          2 allocs/op
BenchmarkLinked_write/1x1-16                     2184786               473.5 ns/op      2162.52 MB/s         830 B/op          0 allocs/op
BenchmarkLinked_read/1MB-16                        33733             36712 ns/op        28561.85 MB/s          0 B/op          0 allocs/op
BenchmarkLinked_free/1MB-16                        80812             14828 ns/op        70717.88 MB/s          0 B/op          0 allocs/op
PASS
```

```
goos: windows
goarch: amd64
cpu: 11th Gen Intel(R) Core(TM) i7-11700F @ 2.50GHz
BenchmarkSized_want-16          100000000               10.69 ns/op            0 B/op          0 allocs/op
BenchmarkSized_write-16          6078380               198.0 ns/op      41379.36 MB/s          0 B/op          0 allocs/op
BenchmarkSized_read-16           4675914               255.1 ns/op      32111.91 MB/s          0 B/op          0 allocs/op
PASS
```
