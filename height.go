package ash

import (
	"math"
)

const (
	CapLevel = 64
)

var (
	// PValue defines the fixed probability that an element in level i appears in level i+1.
	// Defaults to 1/2, another commonly used value for it is 1/4.
	PValue = 0.5
	// RandomHeightFunc is the function that returns the height for any new element to store in a skip list.
	// It is possible to override this function provided that the return value does not exceed maxLevel.
	RandomHeightFunc func(maxLevel int) int = randomHeight
	probabilities    [CapLevel]uint32
)

func init() {
	probability := 1.0

	for level := 1; level < CapLevel; level++ {
		probabilities[level] = uint32(probability * float64(math.MaxUint32))
		probability *= PValue
	}
}

func randomHeight(maxLevel int) int {
	seed := fastrand()

	height := 1
	for height < maxLevel && seed <= probabilities[height] {
		height++
	}

	return height
}
