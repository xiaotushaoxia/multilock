package multilock

import (
	"testing"
)

func TestFixedRW(t *testing.T) {
	// we only use key 0-4 in testMultiLocker
	var mm = NewFixedRW[int32](0, 1, 2, 3, 4)
	testMultiRWLocker(mm)
}
