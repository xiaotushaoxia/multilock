package multilock

import (
	"testing"
)

func TestMultiRWLock(t *testing.T) {
	var mm = NewRWVariable[int32]()

	testMultiRWLocker(mm)
}
