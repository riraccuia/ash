package ash

import (
	"fmt"

	xxhash "github.com/cespare/xxhash/v2"
)

func HashFunc(key any) uint64 {
	if key == nil {
		return 0
	}

	switch v := key.(type) {
	case []byte:
		return xxhash.Sum64(v)
	case string:
		return xxhash.Sum64String(v)
	case uint64:
		return v
	case int:
		return uint64(v)
	default:
		panic(fmt.Sprintf("invalid key type %T", key))
	}
}
