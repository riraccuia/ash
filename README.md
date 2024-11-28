# Ash

Ash is a concurrent, lock-free **A**tomic **S**kiplist **H**ash map. It is intended as a drop-in replacement for `sync.Map` and is designed to operate concurrently without relying on locks, instead using atomic operations and pointer tagging techniques.

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
    m := &ash.Map{}.From(ash.NewSkipList(32))

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

Consider the below representation of a pointer in modern architectures:

```ascii
  63              48  47                                                3     0
 +-+---------------+---+------------------------------------------------+-----+-+
 | 00000000 00000000 | -------- -------- -------- -------- -------- ----- | 000 |
 +------------------------------------------------------------------------------+
High                 |                                                    |    Low
 |                   +--> 3-47 (48 bits) Memory Address (pointer)         |                      
 +--> 48-63 (16 bits) Reserved                                            |
                                                0-2 (3 bits) Alignment <--+
```
Armed with this knowledge, we know that memory address (pointer) representations only use the lower 48 bits of a uint64, which gives us some options to improve logical deletion of nodes from the tree (markable reference).
This package currently encodes a deletion flag (mark) in the lower 4 bits of the top byte (bits 56-59).
Note that once marked, the pointer becomes unusable, and the object will be collected by GC, but that's what we want.


## Contributing

Contributions are encouraged. Issues and pull requests can be submitted through the [GitHub repository](https://github.com/riraccuia/ash).

## License

This code is distributed under the MIT License. See the [LICENSE]([LICENSE](https://github.com/riraccuia/ash/blob/main/LICENSE)) file for details.
