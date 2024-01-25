package multilock

import (
	"sync"
)

type refMutex struct {
	sync.Mutex
	refCounter
}

type defaultLockerFactory struct {
}

func (d *defaultLockerFactory) Get() refTryLocker {
	return &refMutex{}
}

func (d *defaultLockerFactory) Put(locker refTryLocker) {
	return
}

func newPooledLockerFactory() *pooledLockerFactory {
	return &pooledLockerFactory{
		pool: sync.Pool{
			New: func() any {
				return &refMutex{}
			},
		},
	}
}

type pooledLockerFactory struct {
	pool sync.Pool
}

func (d *pooledLockerFactory) Get() refTryLocker {
	return d.pool.Get().(*refMutex)
}

func (d *pooledLockerFactory) Put(locker refTryLocker) {
	d.pool.Put(locker)
	return
}
