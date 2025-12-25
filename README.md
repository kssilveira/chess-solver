# chess-solver

Solver for the [tinyhouse chess variant](https://www.chess.com/variants/tinyhouse).

## Benchmark

```bash
$ go test ./core -bench=. -benchmem -count=3 > core/testdata/benchmark.txt
```
