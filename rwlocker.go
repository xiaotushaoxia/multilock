package multilock

import (
	"sync"
)

type refRWMutex struct {
	sync.RWMutex
	refCounter
}

// defaultRWLockerFactory 有的人会用 sync.Pool 但我测试速度和内存似乎用 sync.Pool都没有优势
type defaultRWLockerFactory struct {
}

func (d *defaultRWLockerFactory) Get() refTryRWLocker {
	return &refRWMutex{}
}

func (d *defaultRWLockerFactory) Put(locker refTryRWLocker) {
	return
}

func newPooledRWLockerFactory() *pooledRWLockerFactory {
	return &pooledRWLockerFactory{
		pool: sync.Pool{
			New: func() any {
				return &refRWMutex{}
			},
		},
	}
}

type pooledRWLockerFactory struct {
	pool sync.Pool
}

func (d *pooledRWLockerFactory) Get() refTryRWLocker {
	return d.pool.Get().(*refRWMutex)
}

func (d *pooledRWLockerFactory) Put(locker refTryRWLocker) {
	d.pool.Put(locker)
	return
}
