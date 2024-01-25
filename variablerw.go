package multilock

import "sync"

func NewRWVariable[K comparable]() MultiRWLocker[K] {
	m := &variableRW[K]{f: &defaultRWLockerFactory{}, locks: map[K]refTryRWLocker{}}
	return m
}

type variableRW[K comparable] struct {
	f rwLockerFactory

	// 最初的版本是没有m的 然后unlock也没加锁。测试发现有bug
	// 所以就把sync.Map换成了普通map，然后map操作的时候加锁
	m     sync.Mutex
	locks map[K]refTryRWLocker
}

func (ml *variableRW[K]) TryRLockByKey(k K) (got bool) {
	_, got = ml.tryRLock(k)
	return
}

func (ml *variableRW[K]) RLockByKey(k K) {
	ml.getLockerAndIncRefCount(k).RLock()
}

func (ml *variableRW[K]) RUnlockByKey(k K) {
	ml.m.Lock()
	l, ok := ml.locks[k]
	ml.m.Unlock()
	if !ok {
		panic("sync: Unlock of unlocked RWMutex")
	}
	ml.rUnlock(k, l)()
}

func (ml *variableRW[K]) TryLockByKey(k K) (got bool) {
	_, got = ml.tryLock(k)
	return
}

func (ml *variableRW[K]) LockByKey(k K) {
	ml.getLockerAndIncRefCount(k).Lock()
}

func (ml *variableRW[K]) UnlockByKey(k K) {
	ml.m.Lock()
	l, ok := ml.locks[k]
	ml.m.Unlock()
	if !ok {
		panic("sync: Unlock of unlocked RWMutex")
	}
	ml.unlock(k, l)()
}

func (ml *variableRW[K]) TryRLock(k K) (got bool, unlock func()) {
	l, got := ml.tryRLock(k)
	if got {
		return true, ml.rUnlock(k, l)
	}
	return false, noop
}

func (ml *variableRW[K]) RLock(k K) (unlock func()) {
	l := ml.getLockerAndIncRefCount(k)
	l.RLock()
	return ml.rUnlock(k, l)

}

func (ml *variableRW[K]) TryLock(k K) (got bool, unlock func()) {
	l, got := ml.tryLock(k)
	if got {
		return true, ml.unlock(k, l)
	}
	return false, noop
}

func (ml *variableRW[K]) Lock(k K) (unlock func()) {
	l := ml.getLockerAndIncRefCount(k)
	l.Lock()
	return ml.unlock(k, l)
}

func (ml *variableRW[K]) tryLock(k K) (l refTryRWLocker, got bool) {
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

func (ml *variableRW[K]) tryRLock(k K) (l refTryRWLocker, got bool) {
	ml.withMutex(func() {
		l = ml._getLockerAndIncRefCount(k)
		if l.TryRLock() {
			got = true
			return
		}
		ml.tryPut(k, l)
		got = false
	})
	return
}

func (ml *variableRW[K]) unlock(k K, l refTryRWLocker) func() {
	// 必须要锁 不然这两个操作不原子 不正确
	return func() {
		ml.withMutex(func() {
			l.Unlock()
			ml.tryPut(k, l)
		})
	}
}

func (ml *variableRW[K]) rUnlock(k K, l refTryRWLocker) func() {
	return func() {
		ml.withMutex(func() {
			l.RUnlock()
			ml.tryPut(k, l)
		})
	}
}

func (ml *variableRW[K]) tryPut(k K, m refTryRWLocker) {
	v := m.DecRefCount()
	if v == 0 {
		delete(ml.locks, k)
		ml.f.Put(m)
	}
}

func (ml *variableRW[K]) tryPutWithMutex(k K, m refTryRWLocker) {
	ml.withMutex(func() {
		ml.tryPut(k, m)
	})
}

func (ml *variableRW[K]) getLockerAndIncRefCount(k K) (l refTryRWLocker) {
	ml.withMutex(func() {
		l = ml._getLockerAndIncRefCount(k)
	})
	return
}

func (ml *variableRW[K]) withMutex(f func()) {
	ml.m.Lock()
	defer ml.m.Unlock()
	f()
}

func (ml *variableRW[K]) _getLockerAndIncRefCount(k K) refTryRWLocker {
	if ml.f == nil {
		ml.f = &defaultRWLockerFactory{}
	}
	if ml.locks == nil {
		ml.locks = map[K]refTryRWLocker{}
	}
	locker, ok := ml.locks[k]
	if !ok {
		locker = ml.f.Get()
		ml.locks[k] = locker
	}
	locker.IncRefCount()
	return locker
}
