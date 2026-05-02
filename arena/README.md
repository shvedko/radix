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
