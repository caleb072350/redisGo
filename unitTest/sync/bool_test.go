package sync_test

import (
	atmc "redisGo/lib/sync/atomic"
	"sync"
	"testing"
)

func TestAtomicBool(t *testing.T) {
	b := new(atmc.AtomicBool)

	// 测试初始状态
	if b.Get() {
		t.Error("Initial value should be false")
	}

	// 测试Set(true)
	b.Set(true)
	if !b.Get() {
		t.Error("Value should be true")
	}

	// 测试Set(false)
	b.Set(false)
	if b.Get() {
		t.Error("Value should be false")
	}

	// 测试并发场景
	var wg sync.WaitGroup
	const numGoroutines = 100
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Set(true)
			if !b.Get() {
				t.Errorf("Concurrent Set(true) failed")
			}
			b.Set(false)
			if b.Get() {
				t.Errorf("Concurrent Set(false) failed")
			}
		}()
	}
	wg.Wait()
}
