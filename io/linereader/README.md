# LineReader
## Usage
```go
import "github.com/asymmetric-research/go-commons/io/linereader"

lr := linereader.New(reader, 4096 /* blockSize */)
line := [12288]byte{}

var err error
for err == nil {
    n, ntrunc, err := lr.Read(line[:])
    lastline := line[:n]
    fmt.Println("%d bytes didn't fit", ntrunc)
}
```

## Benchmarks
```
go test -benchmem -benchtime=5s -bench=. ./io/linereader/...
goos: linux
goarch: amd64
pkg: github.com/asymmetric-research/go-commons/io/linereader
cpu: AMD Ryzen 9 5950X 16-Core Processor
BenchmarkLineReaderUnbuffered-32         1241234              4680 ns/op           22560 B/op          5 allocs/op
BenchmarkHashicorpsUnbuffered-32            4722           1349399 ns/op         2295410 B/op      29602 allocs/op
BenchmarkGoCmdUnbuffered-32               234085             24217 ns/op           41636 B/op        289 allocs/op
BenchmarkLineReaderLargeReads-32         2210827              2713 ns/op           12328 B/op          4 allocs/op
BenchmarkHashicorpsLargeReads-32            4208           1406119 ns/op         2285073 B/op      29601 allocs/op
BenchmarkGoCmdLargeReads-32               274774             21724 ns/op           31563 B/op        292 allocs/op
```