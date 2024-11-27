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

## Contributing

Contributions are encouraged. Issues and pull requests can be submitted through the [GitHub repository](https://github.com/riraccuia/ash).

## License

This code is distributed under the MIT License. See the [LICENSE]([LICENSE](https://github.com/riraccuia/ash/blob/main/LICENSE)) file for details.
