package multilock

import (
	"testing"
)

func TestFixed(t *testing.T) {
	// we only use key 0-4 in testMultiLocker
	var mm = NewFixed[int32](0, 1, 2, 3, 4)
	testMultiLocker(mm)
}
