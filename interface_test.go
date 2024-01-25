package multilock

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

func testMultiRWLocker(mm MultiRWLocker[int32]) map[int32]map[int32]int32 {
	mp := createTestMap()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_testMultiLocker(mm, &wg, mp)
		wg.Done()
	}()
	var vv atomic.Int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		ii := int32(i)
		go func() {
			defer wg.Done()

			mod := ii % 4
			switch mod {
			case 0:
				unlock := mm.RLock(mod)
				vv.Store(mp[mod][ii])
				unlock()
			case 1:
				ok, unlock := mm.TryRLock(mod)
				if ok {
					vv.Store(mp[mod][ii])
				}
				unlock()
			case 2:
				mm.RLockByKey(mod)
				//mp[mod][ii] = ii // will panic: concurrent map writes
				vv.Store(mp[mod][ii])
				mm.RUnlockByKey(mod)
			case 3:
				mm.TryRLockByKey(mod)
				if mm.TryRLockByKey(mod) {
					vv.Store(mp[mod][ii])
					mm.RUnlockByKey(mod)
				}
			}
		}()
	}
	wg.Wait()
	return mp
}

func testMultiLocker(mm MultiLocker[int32]) map[int32]map[int32]int32 {
	var wg sync.WaitGroup
	mp := createTestMap()
	_testMultiLocker(mm, &wg, mp)
	wg.Wait()
	return mp
}

func _testMultiLocker(mm MultiLocker[int32], wg *sync.WaitGroup, mp map[int32]map[int32]int32) {
	var vv atomic.Int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		ii := int32(i)
		go func() {
			defer wg.Done()
			mod := ii % 4
			switch mod {
			case 0:
				unlock := mm.Lock(mod)
				mp[mod][ii] = ii
				unlock()
			case 1:
				ok, unlock := mm.TryLock(mod)
				if ok {
					mp[mod][ii] = ii
				}
				unlock()
			case 2:
				mm.LockByKey(mod)
				vv.Store(mp[mod][ii])
				mm.UnlockByKey(mod)
			case 3:
				if mm.TryLockByKey(mod) {
					vv.Store(mp[mod][ii])
					mm.UnlockByKey(mod)
				}
			}
		}()
	}
}

func createTestMap() map[int32]map[int32]int32 {
	var mp = map[int32]map[int32]int32{}
	mp[0] = map[int32]int32{}
	mp[1] = map[int32]int32{}
	mp[2] = map[int32]int32{}
	mp[3] = map[int32]int32{}
	mp[4] = map[int32]int32{}
	return mp
}

func mapEq(m1, m2 map[int32]map[int32]int32) bool {
	marshal1, err := json.Marshal(m1)
	if err != nil {
		panic(err)
	}
	marshal2, err := json.Marshal(m2)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshal1))
	fmt.Println(string(marshal2))
	if len(marshal2) != len(marshal1) {
		return false
	}
	for i, b := range marshal2 {
		if b != marshal1[i] {
			return false
		}
	}
	return true
}
