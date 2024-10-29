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

func (lockMap *LockMap) toLockIndices(keys []string, reverse bool) []uint32 {
	indexMap := make(map[uint32]bool)
	for _, key := range keys {
		index := lockMap.spread(fnv32(key))
		indexMap[index] = true
	}
	indices := make([]uint32, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return indices[i] < indices[j]
		} else {
			return indices[i] > indices[j]
		}
	})
	return indices
}

func (lockMap *LockMap) Locks(keys ...string) {
	indices := lockMap.toLockIndices(keys, false)
	for _, index := range indices {
		mu := lockMap.table[index]
		mu.Lock()
	}
}

func (lockMap *LockMap) Unlocks(keys ...string) {
	indices := lockMap.toLockIndices(keys, true)
	for _, index := range indices {
		mu := lockMap.table[index]
		mu.Unlock()
	}
}

func (lockMap *LockMap) RLocks(keys ...string) {
	indices := lockMap.toLockIndices(keys, false)
	for _, index := range indices {
		mu := lockMap.table[index]
		mu.RLock()
	}
}

func (lockMap *LockMap) RUnlocks(keys ...string) {
	indices := lockMap.toLockIndices(keys, true)
	for _, index := range indices {
		mu := lockMap.table[index]
		mu.RUnlock()
	}
}
