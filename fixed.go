package multilock

import (
	"fmt"
	"sync"
)

func NewFixed[K comparable](keys ...K) MultiLocker[K] {
	f := fixed[K]{}
	for _, key := range keys {
		f[key] = &sync.Mutex{}
	}
	return f
}

type fixed[K comparable] map[K]*sync.Mutex

//type fixed[K comparable] struct {
//	ms map[K]*sync.Mutex
//}

func (f fixed[K]) TryLockByKey(k K) bool {
	return f.mustGet(k).TryLock()
}

func (f fixed[K]) LockByKey(k K) {
	f.mustGet(k).Lock()
}

func (f fixed[K]) UnlockByKey(k K) {
	f.mustGet(k).Unlock()
}

func (f fixed[K]) TryLock(k K) (got bool, unlock func()) {
	l := f.mustGet(k)
	got = l.TryLock()
	if got {
		unlock = func() { l.Unlock() }
	} else {
		unlock = noop
	}
	return
}

func (f fixed[K]) Lock(k K) (unlock func()) {
	l := f.mustGet(k)
	l.Lock()
	return func() { l.Unlock() }
}

func (f fixed[K]) mustGet(k K) *sync.Mutex {
	//mutex := f.ms[k]
	mutex := f[k]
	if mutex == nil {
		panic(fmt.Errorf("%w: %v", ErrLockKeyNotFound, k))
	}
	return mutex
}
