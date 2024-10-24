package lock

import (
	"sort"
	"sync"
)

const (
	prime32 = uint32(16777619)
)

type LockMap struct {
	table []*sync.RWMutex
}

func Make(size int) *LockMap {
	table := make([]*sync.RWMutex, size)
	for i := 0; i < size; i++ {
		table[i] = &sync.RWMutex{}
	}

	return &LockMap{table: table}
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (lockMap *LockMap) spread(hashCode uint32) uint32 {
	if lockMap == nil {
		panic("lockMap is nil")
	}
	size := uint32(len(lockMap.table))
	return (size - 1) & hashCode
}

func (lockMap *LockMap) Lock(key string) {
	index := lockMap.spread(fnv32(key))
	mu := lockMap.table[index]
	mu.Lock()
}

func (lockMap *LockMap) Unlock(key string) {
	index := lockMap.spread(fnv32(key))
	mu := lockMap.table[index]
	mu.Unlock()
}

func (lockMap *LockMap) RLock(key string) {
	index := lockMap.spread(fnv32(key))
	mu := lockMap.table[index]
	mu.RLock()
}

func (lockMap *LockMap) RUnlock(key string) {
	index := lockMap.spread(fnv32(key))
	mu := lockMap.table[index]
	mu.RUnlock()
}

func (lockMap *LockMap) Locks(keys ...string) {
	keySlice := make(sort.StringSlice, len(keys))
	copy(keySlice, keys)
	sort.Sort(keySlice)
	for _, key := range keySlice {
		lockMap.Lock(key)
	}
}

func (lockMap *LockMap) Unlocks(keys ...string) {
	size := len(keys)
	keySlice := make(sort.StringSlice, size)
	copy(keySlice, keys)
	sort.Sort(keySlice)
	for i := size - 1; i >= 0; i-- {
		key := keySlice[i]
		lockMap.Unlock(key)
	}
}

func (lockMap *LockMap) RLocks(keys ...string) {
	keySlice := make(sort.StringSlice, len(keys))
	copy(keySlice, keys)
	sort.Sort(keySlice)
	for _, key := range keySlice {
		lockMap.RLock(key)
	}
}

func (lockMap *LockMap) RUnlocks(keys ...string) {
	size := len(keys)
	keySlice := make(sort.StringSlice, size)
	copy(keySlice, keys)
	sort.Sort(keySlice)
	for i := size - 1; i >= 0; i-- {
		key := keySlice[i]
		lockMap.RUnlock(key)
	}
}
