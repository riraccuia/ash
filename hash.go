package ash

import (
	xxhash "github.com/cespare/xxhash/v2"
)

func HashFunc(key any) uint64 {
	if key == nil {
		return 0
	}

	var _key []byte
	switch v := key.(type) {
	case []byte:
		_key = v
	case string:
		_key = []byte(v)
	case int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		_key = NumericToBytesUnsafe(v)
	default:
		panic("invalid key type")
	}
	return xxhash.Sum64(_key)
}
