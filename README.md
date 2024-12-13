# Ash
[![Go Reference](https://pkg.go.dev/badge/github.com/riraccuia/ash.svg)](https://pkg.go.dev/github.com/riraccuia/ash)

Ash is a concurrent, lock-free **A**tomic **S**kiplist **H**ash map written in golang. It is intended as a drop-in replacement for `sync.Map` and is designed to operate concurrently without relying on locks, instead using atomic operations and pointer tagging techniques.

## Motivation

As part of writing go code professionally, built-in `map` and `sync.Map` are certainly among the data structures that I have used the most, and not without some level of frustration.  
`sync.Map` was a great addition to the standard library, but it's not ideal in every scenario: more often than not I find myself guarding a map's reads and writes using synchronization primitives like `Mutex` and/or `RWMutex` or some combination of both.  
YunHao Zhang's talk<sup>[[1]]</sup> at Gophercon 2024 in Chicago was an inspiring introduction to the **skip list** data structure, which surprisingly enough is not implemented in any of the go standard lib.  
This library explores a lock-free implementation of the skip list as a way to challenge myself and learn something new.
See the [References](https://github.com/riraccuia/ash?tab=readme-ov-file#pointer-tagging) section for the list of papers that inspired the implementation.

## Features

- **SkipList based design**: uses a skip list as the underlying data structure, allowing efficient insertion, deletion, and search operations.
- **Concurrent**: lock-free atomic implementation
- **Pointer Tagging**: employs pointer tagging techniques to achieve a markable reference that encodes metadata (a mark bit) directly within pointer values, facilitating atomic updates to the structure.
- **Compatibility**: consistent with `sync.Map`, straightforward adoption in existing codebases.

## Status

This library is under active development.  
At this time I have only tested on the following:
```
goos: darwin
goarch: arm64
pkg: ash
cpu: Apple M1
```

## Usage

The usage of `ash.Map` mirrors that of `sync.Map`. The following example demonstrates basic operations:

```go
package main

import (
    "fmt"
    "github.com/riraccuia/ash"
)

func main() {
    m := new(ash.Map).From(ash.NewSkipList(32))

    // Store a key-value pair
    m.Store("key", "value")

    // Retrieve a value
    value, ok := m.Load("key")
    if ok {
        fmt.Println("Loaded:", value)
    }

    // Delete a key-value pair
    m.Delete("key")
}
```
## Pointer Tagging

Consider the below representation of a pointer in modern architectures<sup>[[2]]</sup>:

```ascii

                                                0-2 (3 bits) Alignment <--+
                                                                          |
                                                                          |
  63              48  47                                                3 |   0
 +-+---------------+---+------------------------------------------------+-----+-+
 | 00000000 00000000 | -------- -------- -------- -------- -------- ----- | 000 |
 +------------------------------------------------------------------------------+
High                 |                                                         Low
 |                   +--> 0-47 (48 bits) Memory Address (pointer)
 |
 +--> 48-63 (16 bits) Reserved
```
Armed with this knowledge, we know that memory address (pointer) representations only use the lower 48 bits of a uint64, which gives us some options to improve logical deletion of nodes from the tree (markable reference).
This package currently encodes a deletion flag (mark) in the top byte (bits 56-63) of the address.  
TBI<sup>[[3]]</sup> (Top Byte Ignore) should allow direct usage of the tainted pointer on linux/macOS with `aarch64` but the top 8 bits of the address are being cleared prior to consuming it.

## Benchmarks

```
% go test -v -run=NOTEST -bench=. -benchtime=5000000x -benchmem -cpu=8 -count=1
goos: darwin
goarch: arm64
pkg: ash
cpu: Apple M1
BenchmarkSyncMap_70Load20Store10Delete
    map_bench_test.go:50: sync.Map total calls to Store/Delete/Load:  0 / 0 / 1 /
    map_bench_test.go:55: Execution time:  62.917µs
    map_bench_test.go:50: sync.Map total calls to Store/Delete/Load:  1000010 / 499183 / 3500807 /
    map_bench_test.go:55: Execution time:  1.415291833s
BenchmarkSyncMap_70Load20Store10Delete-8         5000000               283.1 ns/op            52 B/op          0 allocs/op
BenchmarkAshMap_70Load20Store10Delete
    map_bench_test.go:97: ash.Map total calls to Store/Delete/Load:  0 / 0 / 1 /
    map_bench_test.go:103: Execution time:  10.75µs
    map_bench_test.go:97: ash.Map total calls to Store/Delete/Load:  1000134 / 499997 / 3499869 /
    map_bench_test.go:103: Execution time:  484.099ms
BenchmarkAshMap_70Load20Store10Delete-8          5000000                96.82 ns/op           84 B/op          1 allocs/op
```

## Contributing

Contributions are encouraged. Issues and pull requests can be submitted through the [GitHub repository](https://github.com/riraccuia/ash).

## License

This code is distributed under the MIT License. See the [LICENSE](https://github.com/riraccuia/ash/blob/main/LICENSE) file for details.

## Credits

- https://github.com/cespare/xxhash : great Go implementation of the 64-bit xxHash algorithm, this is what ash uses to hash keys.
- https://www.cloudcentric.dev/implementing-a-skip-list-in-go/ : pseudo-random height generation.
- https://github.com/zhangyunhao116/skipmap : YunHao's work, thanks for the inspiration.

## References

- https://supertaunt.github.io/CMU_15618_project.github.io/
- https://homepage.cs.uiowa.edu/%7Eghosh/skip.pdf

[1]: https://github.com/gophercon/2024-talks/tree/main/YunHaoZhang-BuildingaHighPerformanceConcurrentMapInGo "Building a High Performace Concurrent Map In Go {YunHao Zhang}"
[2]: https://dl.acm.org/doi/abs/10.1145/3558200 "A Primer on Pointer Tagging {Chaitanya Koparkar}"
[3]: https://www.linaro.org/blog/top-byte-ignore-for-fun-and-memory-savings/ "Top Byte Ignore For Fun and Memory Savings"
