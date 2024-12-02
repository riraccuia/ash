# Ash

Ash is a concurrent, lock-free **A**tomic **S**kiplist **H**ash map written in golang. It is intended as a drop-in replacement for `sync.Map` and is designed to operate concurrently without relying on locks, instead using atomic operations and pointer tagging techniques.

## Motivation

As part of writing go code professionally, built-in `map` and `sync.Map` are certainly among the data structures that I have used the most, and not without some level of frustration.  
`sync.Map` was a great addition to the standard library, but it's not ideal in every scenario: more often than not you just might be better off guarding a map's reads and writes using synchronization primitives like `Mutex` and/or `RWMutex` or some combination of both.  
With caches and network/priority queues being an important part of the codebases that I maintain at my company, I love learning new ways to squeeze a little performance.  
YunHao Zhang's talk<sup>[[1]]</sup> at Gophercon 2024 in Chicago was an inspiring introduction to the **skip list** data structure, which surprisingly enough is not implemented in any of the go standard lib.  
This library explores a lock-free implementation of the skip list as a way for me to both intimately understand it and get more hands-on experience with atomic primitives, as well as become a base for some future work.
See the [References](https://github.com/riraccuia/ash?tab=readme-ov-file#pointer-tagging) section for the list of papers that inspired the implementation.

## Features

- **SkipList based design**: uses a skip list as the underlying data structure, allowing efficient insertion, deletion, and search operations.
- **Concurrent**: lock-free atomic implementation
- **Pointer Tagging**: employs pointer tagging techniques to achieve a markable reference that encodes metadata (a mark bit) directly within pointer values, facilitating atomic updates to the structure.
- **Compatibility**: consistent with `sync.Map`, straightforward adoption in existing codebases.

## Status

This library is under active development

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
This package currently encodes a deletion flag (mark) in the lower 4 bits of the top byte (bits 56-59).  
Note that this currently works out of the box on `aarch64` thanks to TBI<sup>[[3]]</sup> (Top Byte Ignore) on linux and macOS (apple silicon).  
For `amd64` the plan is to use `runtime.SetFinalizer` and `runtime.KeepAlive` to prevent GC from collecting Nodes too soon and crash while walking the tree.

## Contributing

Contributions are encouraged. Issues and pull requests can be submitted through the [GitHub repository](https://github.com/riraccuia/ash).

## License

This code is distributed under the MIT License. See the [LICENSE](https://github.com/riraccuia/ash/blob/main/LICENSE) file for details.

## References

- https://supertaunt.github.io/CMU_15618_project.github.io/
- https://homepage.cs.uiowa.edu/%7Eghosh/skip.pdf

[1]: https://github.com/gophercon/2024-talks/tree/main/YunHaoZhang-BuildingaHighPerformanceConcurrentMapInGo "Building a High Performace Concurrent Map In Go {YunHao Zhang}"
[2]: https://dl.acm.org/doi/abs/10.1145/3558200 "A Primer on Pointer Tagging {Chaitanya Koparkar}"
[3]: https://www.linaro.org/blog/top-byte-ignore-for-fun-and-memory-savings/ "Top Byte Ignore For Fun and Memory Savings"
