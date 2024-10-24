package dict

import (
	"redisGo/interface/dict"
	"sync"
	"sync/atomic"
)

type ConcurrentDict struct {
	table []*Shard
	count int32
}

type Shard struct {
	m     map[string]interface{}
	mutex sync.RWMutex
}

const (
	maxCapacity = 1 << 16
	minCapacity = 256
)

func computeCapacity(param int) (size int) {
	if param <= minCapacity {
		return minCapacity
	}
	n := param - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return maxCapacity
	}
	return n + 1
}

func MakeConcurrent(shardCountHint int) *ConcurrentDict {
	shardCount := computeCapacity(shardCountHint)
	shards := make([]*Shard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &Shard{m: make(map[string]interface{})}
	}
	return &ConcurrentDict{table: shards, count: 0}
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (d *ConcurrentDict) spread(key string) int {
	h := int(fnv32(key))
	size := len(d.table)
	return (size - 1) & h
}

func (d *ConcurrentDict) Get(key string) (val interface{}, exists bool) {
	shard := d.table[d.spread(key)]
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	val, ok := shard.m[key]
	return val, ok
}

func (d *ConcurrentDict) Len() int {
	return int(atomic.LoadInt32(&d.count))
}

func (d *ConcurrentDict) Put(key string, val interface{}) int {
	shard := d.table[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.m[key]
	if exists {
		return 0
	} else {
		shard.m[key] = val
		atomic.AddInt32(&d.count, 1)
		return 1
	}
}

func (d *ConcurrentDict) PutIfAbsent(key string, val interface{}) int {
	shard := d.table[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.m[key]
	if exists {
		return 0
	} else {
		shard.m[key] = val
		return 1
	}
}

func (d *ConcurrentDict) PutIfExists(key string, val interface{}) int {
	shard := d.table[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.m[key]
	if exists {
		shard.m[key] = val
		return 1
	} else {
		return 0
	}
}

func (d *ConcurrentDict) Remove(key string) int {
	shard := d.table[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.m[key]
	if exists {
		delete(shard.m, key)
		atomic.AddInt32(&d.count, -1)
		return 1
	} else {
		return 0
	}
}

func (d *ConcurrentDict) ForEach(consumer dict.Consumer) {
	for _, shard := range d.table {
		shard.mutex.RLock()
		for k, v := range shard.m {
			if !consumer(k, v) {
				break
			}
		}
		shard.mutex.RUnlock()
	}
}
