package dict

import (
	"sync"
	"sync/atomic"
)

type Dict struct {
	shards     []*Shard
	shardCount int
	count      int32
}

type Shard struct {
	table map[string]interface{}
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

func Make(shardCountHint int) *Dict {
	shardCount := computeCapacity(shardCountHint)
	shards := make([]*Shard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &Shard{table: make(map[string]interface{})}
	}
	return &Dict{shards: shards, shardCount: shardCount}
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

func (d *Dict) spread(key string) int {
	h := int(fnv32(key))
	return (d.shardCount - 1) & h
}

func (d *Dict) Get(key string) (val interface{}, exists bool) {
	shard := d.shards[d.spread(key)]
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	val, ok := shard.table[key]
	return val, ok
}

func (d *Dict) Len() int {
	return int(atomic.LoadInt32(&d.count))
}

func (d *Dict) Put(key string, val interface{}) int {
	shard := d.shards[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.table[key]
	if exists {
		return 0
	} else {
		shard.table[key] = val
		atomic.AddInt32(&d.count, 1)
		return 1
	}
}

func (d *Dict) PutIfAbsent(key string, val interface{}) int {
	shard := d.shards[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.table[key]
	if exists {
		return 0
	} else {
		shard.table[key] = val
		return 1
	}
}

func (d *Dict) PutIfExists(key string, val interface{}) int {
	shard := d.shards[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.table[key]
	if exists {
		shard.table[key] = val
		return 1
	} else {
		return 0
	}
}

func (d *Dict) Remove(key string) int {
	shard := d.shards[d.spread(key)]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.table[key]
	if exists {
		delete(shard.table, key)
		atomic.AddInt32(&d.count, -1)
		return 1
	} else {
		return 0
	}
}
