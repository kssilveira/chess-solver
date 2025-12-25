# chess-solver

Solver for the [tinyhouse chess variant](https://www.chess.com/variants/tinyhouse).

## Benchmark

```bash
$ go test ./core -bench=. -benchmem -count=6 > core/testdata/benchmark.txt
$ $GOPATH/bin/benchstat <(git show main:core/testdata/benchmark.txt) core/testdata/benchmark.txt > core/testdata/benchstat.txt
```
