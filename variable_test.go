package multilock

import "testing"

func TestMultiLock(t *testing.T) {
	var mm = NewVariable[int32]()

	testMultiLocker(mm)
}
