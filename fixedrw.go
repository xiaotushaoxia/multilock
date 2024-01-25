package multilock

import (
	"fmt"
	"sync"
)

func NewFixedRW[K comparable](keys ...K) MultiRWLocker[K] {
	//ms := map[K]*sync.RWMutex{}
	//for _, key := range keys {
	//	ms[key] = &sync.RWMutex{}
	//}
	//return &fixedRW[K]{ms: ms}

	f := fixedRW[K]{}
	for _, key := range keys {
		f[key] = &sync.RWMutex{}
	}
	return f
}

type fixedRW[K comparable] map[K]*sync.RWMutex

//type fixedRW[K comparable] struct {
//	ms map[K]*sync.RWMutex
//}

func (f fixedRW[K]) TryRLockByKey(k K) bool {
	return f.mustGet(k).TryRLock()
}

func (f fixedRW[K]) RLockByKey(k K) {
	f.mustGet(k).RLock()
}

func (f fixedRW[K]) RUnlockByKey(k K) {
	f.mustGet(k).RUnlock()
}

func (f fixedRW[K]) TryRLock(k K) (got bool, unlock func()) {
	l := f.mustGet(k)
	got = l.TryRLock()
	if got {
		unlock = func() { l.RUnlock() }
	} else {
		unlock = noop
	}
	return
}

func (f fixedRW[K]) RLock(k K) (unlock func()) {
	l := f.mustGet(k)
	l.RLock()
	return func() { l.RUnlock() }
}

func (f fixedRW[K]) TryLockByKey(k K) bool {
	return f.mustGet(k).TryLock()
}

func (f fixedRW[K]) LockByKey(k K) {
	f.mustGet(k).Lock()
}

func (f fixedRW[K]) UnlockByKey(k K) {
	f.mustGet(k).Unlock()
}

func (f fixedRW[K]) TryLock(k K) (got bool, unlock func()) {
	l := f.mustGet(k)
	got = l.TryLock()
	if got {
		unlock = func() { l.Unlock() }
	} else {
		unlock = noop
	}
	return
}

func (f fixedRW[K]) Lock(k K) (unlock func()) {
	l := f.mustGet(k)
	l.Lock()
	return func() { l.Unlock() }
}

func (f fixedRW[K]) mustGet(k K) *sync.RWMutex {
	//mutex := f.ms[k]
	mutex := f[k]
	if mutex == nil {
		panic(fmt.Errorf("%w: %v", ErrLockKeyNotFound, k))
	}
	return mutex
}
