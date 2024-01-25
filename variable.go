package multilock

import (
	"sync"
)

func NewVariable[K comparable]() MultiLocker[K] {
	m := &variable[K]{f: &defaultLockerFactory{}, locks: map[K]refTryLocker{}}
	return m
}

type variable[K comparable] struct {
	f lockerFactory

	m     sync.Mutex
	locks map[K]refTryLocker
}

func (ml *variable[K]) TryLockByKey(k K) (got bool) {
	_, got = ml.tryLock(k)
	return
}

func (ml *variable[K]) LockByKey(k K) {
	ml.getLockerAndIncRefCount(k).Lock()
}

func (ml *variable[K]) UnlockByKey(k K) {
	ml.m.Lock()
	l, ok := ml.locks[k]
	ml.m.Unlock()
	if !ok {
		panic("sync: unlock of unlocked mutex")
	}
	ml.unlock(k, l)()
}

func (ml *variable[K]) TryLock(k K) (got bool, unlock func()) {
	l, got := ml.tryLock(k)
	if got {
		return true, ml.unlock(k, l)
	}
	return false, noop
}

func (ml *variable[K]) Lock(k K) (unlock func()) {
	l := ml.getLockerAndIncRefCount(k)
	l.Lock()
	return ml.unlock(k, l)
}

func (ml *variable[K]) tryLock(k K) (l refTryLocker, got bool) {
	ml.withMutex(func() {
		l = ml._getLockerAndIncRefCount(k)
		if l.TryLock() {
			got = true
			return
		}
		ml.tryPut(k, l)
		got = false
	})
	return
}

func (ml *variable[K]) unlock(k K, l refTryLocker) func() {
	return func() {
		ml.withMutex(func() {
			ml.tryPut(k, l)
			l.Unlock()
		})
	}
}

func (ml *variable[K]) tryPut(k K, m refTryLocker) {
	v := m.DecRefCount()
	if v == 0 {
		delete(ml.locks, k)
		ml.f.Put(m)
	}
}

func (ml *variable[K]) tryPutWithMutex(k K, m refTryLocker) {
	ml.withMutex(func() {
		ml.tryPut(k, m)
	})
}

func (ml *variable[K]) getLockerAndIncRefCount(k K) (l refTryLocker) {
	ml.withMutex(func() {
		l = ml._getLockerAndIncRefCount(k)
	})
	return
}

func (ml *variable[K]) withMutex(f func()) {
	ml.m.Lock()
	defer ml.m.Unlock()
	f()
}

func (ml *variable[K]) _getLockerAndIncRefCount(k K) refTryLocker {
	if ml.f == nil {
		ml.f = &defaultLockerFactory{}
	}
	if ml.locks == nil {
		ml.locks = map[K]refTryLocker{}
	}
	locker, ok := ml.locks[k]
	if !ok {
		locker = ml.f.Get()
		ml.locks[k] = locker
	}
	locker.IncRefCount()
	return locker
}
