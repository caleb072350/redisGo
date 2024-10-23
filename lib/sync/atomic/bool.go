package atomic

import (
	"sync/atomic"
)

type AtomicBool uint32

// type AtomicBool struct {
// 	value uint32
// 	mu    sync.RWMutex
// }

func (b *AtomicBool) Get() bool {
	// b.mu.RLock()
	// defer b.mu.RUnlock()
	return atomic.LoadUint32((*uint32)(b)) != 0
}

func (b *AtomicBool) Set(v bool) {
	// b.mu.Lock()
	// defer b.mu.Unlock()
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
