# chess-solver

Solver for the [tinyhouse chess variant](https://www.chess.com/variants/tinyhouse).

## Play against the solution for a given board position

```bash
$ go run main.go --board="KRNB,N   ,    ,   k" | tee core/testdata/play.txt
```

See example game on [core/testdata/play.txt](core/testdata/play.txt):

- play starts moving the King (K)

|KRNB|
|N   |
|    |
|   k|
‾‾‾‾‾‾
move: K (0, 0) =>   (1, 1)
______
| RNB|
|NK  |
|    |
|   k|
‾‾‾‾‾‾

- Knight (N) sacrifice
______
|   B|
| KR |
| N  |
| Nk |
‾‾‾‾‾‾
move: K (1, 1) =>   (1, 0)
______
|   B|
|K R |
| N  |
| Nk |
‾‾‾‾‾‾
- Rook (R) sacrifice
______
|   B|
|   R|
|KNk |
|    |
‾‾‾‾‾‾
move: R (1, 3) =>   (2, 3)
______
|   B|
|    |
|KNkR|
|    |
‾‾‾‾‾‾
- king (k) trapped
______
|   B|
|    |
|KN k|
|    |
‾‾‾‾‾‾
move: K (2, 0) =>   (3, 1)
______
|   B|
|    |
| N k|
| K  |
‾‾‾‾‾‾

## Benchmark

```bash
$ time go test ./core -bench=. -benchmem -count=6 -cpuprofile cpu.txt -memprofile mem.txt > core/testdata/benchmark.txt
$ $GOPATH/bin/benchstat <(git show main:core/testdata/benchmark.txt) core/testdata/benchmark.txt > core/testdata/benchstat.txt
$ go tool pprof cpu.txt
$ go tool pprof mem.txt
```

See history of benchmark improvements for [benchmark.txt](commits/main/core/testdata/benchmark.txt) and [benchstat.txt](commits/main/core/testdata/benchstat.txt).
