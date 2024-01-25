package multilock

import (
	"sync"
	"testing"
)

func Test_refCounter(t *testing.T) {
	var d defaultLockerFactory
	get := d.Get()
	var wg sync.WaitGroup
	wg.Add(200)
	for i := 0; i < 100; i++ {
		go func() {
			wg.Done()
			get.IncRefCount()
		}()
	}
	for i := 0; i < 100; i++ {
		go func() {
			wg.Done()
			get.DecRefCount()
		}()
	}
	wg.Wait()

	if got := get.GetRefCount(); got != 0 {
		t.Errorf("GetRefCount() = %v, want %v", got, 0)
	}
}
