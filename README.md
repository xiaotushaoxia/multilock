# multilock

## usage
### 演示了通过multilock对map的并发写入 
```go
package main

import (
	"sync"

	"github.com/xiaotushaoxia/multilock"
)

func main() {
	var ml = multilock.NewVariable[int]()
	
	mp := map[int]map[int]int{}
	count := 1000
	for i := 0; i < count; i++ {
		mp[i%10] = map[int]int{}
	}
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		seg := i % 10
		ii := i
		if ii%4 == 0 {
			go func() {
				defer wg.Done()
				unlock := ml.Lock(seg)
				defer unlock()
				mp[seg][ii] = ii
			}()
		} else if ii%4 == 1 {
			go func() {
				defer wg.Done()
				ml.LockByKey(seg)
				defer ml.UnlockByKey(seg)
				mp[seg][ii] = ii
			}()
		} else if ii%4 == 3 {
			go func() {
				defer wg.Done()
				if got, unlock := ml.TryLock(seg); got {
					defer unlock()
					mp[seg][ii] = ii
				}
			}()
		} else {
			go func() {
				defer wg.Done()
				if ml.TryLockByKey(seg) {
					mp[seg][ii] = ii
					ml.UnlockByKey(seg)
				}
			}()
		}
	}
	wg.Wait()
}

```