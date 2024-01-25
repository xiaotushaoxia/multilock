package multilock

import (
	"fmt"
	"sync/atomic"
)

// public

type MultiLocker[K comparable] interface {
	TryLockByKey(K) bool
	LockByKey(K)
	UnlockByKey(K)

	TryLock(K) (got bool, unlock func())
	Lock(K) (unlock func())
}

type MultiRWLocker[K comparable] interface {
	TryLockByKey(K) bool
	LockByKey(K)
	UnlockByKey(K)

	TryLock(K) (got bool, unlock func())
	Lock(K) (unlock func())

	TryRLockByKey(K) bool
	RLockByKey(K)
	RUnlockByKey(K)

	TryRLock(K) (got bool, unlock func())
	RLock(K) (unlock func())
}

var ErrLockKeyNotFound = fmt.Errorf("locker key not found")

// private
var noop = func() {}

type refTryLocker interface {
	tryLocker
	refCountable
}

type refTryRWLocker interface {
	tryRWLocker
	refCountable
}

type tryLocker interface {
	Lock()
	TryLock() bool
	Unlock()
}

type tryRLocker interface {
	RLock()
	TryRLock() bool
	RUnlock()
}

type tryRWLocker interface {
	tryRLocker
	tryLocker
}

type lockerFactory interface {
	Get() refTryLocker
	Put(refTryLocker)
}

type rwLockerFactory interface {
	Get() refTryRWLocker
	Put(refTryRWLocker)
}

type refCountable interface {
	GetRefCount() int64
	IncRefCount() int64
	DecRefCount() int64
}

// 默认实现
type refCounter struct {
	atomic.Int64
}

func (c *refCounter) GetRefCount() int64 {
	return c.Int64.Load()
}

func (c *refCounter) IncRefCount() int64 {
	return c.Int64.Add(1)
}

func (c *refCounter) DecRefCount() int64 {
	return c.Int64.Add(-1)
}
