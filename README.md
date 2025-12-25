# chess-solver

Solver for the [tinyhouse chess variant](https://www.chess.com/variants/tinyhouse).

## Benchmark

```bash
$ time go test ./core -bench=. -benchmem -count=6 -cpuprofile cpu.txt -memprofile mem.txt > core/testdata/benchmark.txt
$ $GOPATH/bin/benchstat <(git show main:core/testdata/benchmark.txt) core/testdata/benchmark.txt > core/testdata/benchstat.txt
$ go tool pprof cpu.txt
$ go tool pprof mem.txt
```
