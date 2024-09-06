# io

## LineReader
### Usage
```go
lr := NewLineReader(reader, 4096 /* blockSize */)
line := [12288]byte{}

var err error
for err == nil {
    n, ntrunc, err := lr.Read(line[:])
    lastline := line[:n]
    fmt.Println("%d bytes didn't fit", ntrunc)
}
```

### Benchmarks
```
go test -benchmem -benchtime=5s -bench=. ./io/...
goos: linux
goarch: amd64
pkg: github.com/asymmetric-research/go-commons/io
cpu: AMD Ryzen 9 5950X 16-Core Processor
BenchmarkLineReaderUnbuffered-32         1237466              4726 ns/op           22560 B/op          5 allocs/op
BenchmarkHashicorpsUnbuffered-32            4712           1345807 ns/op         2295415 B/op      29602 allocs/op
BenchmarkGoCmdUnbuffered-32               236834             24471 ns/op           41636 B/op        289 allocs/op
BenchmarkLineReader-32                   2206489              2722 ns/op           12328 B/op          4 allocs/op
BenchmarkHashicorps-32                      4239           1416668 ns/op         2285070 B/op      29601 allocs/op
BenchmarkGoCmd-32                         272906             21842 ns/op           31563 B/op        292 allocs/op
PASS
ok      github.com/asymmetric-research/go-commons/io    44.384s
```