package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
)

type HashFunc func(data []byte) uint32

type Map struct {
	hashFunc HashFunc
	replicas int // 表示每个键的副本数量（虚拟节点数量）。一致性哈希为了更好的负载均衡，通常会为每个服务器创建多个虚拟节点。
	// replicas 指定了每个实际服务器在哈希环上对应的虚拟节点数量。值越大，负载均衡效果越好，但同时也增加了计算开销。
	keys []int //sorted 存储所有虚拟节点的哈希值。这些哈希值代表了虚拟节点在哈希环上的位置。这个切片是按照哈希值排序的，可以使用二分查找来
	//高效地查找最近的节点。
	hashMap map[int]string // 键是虚拟节点的哈希值(int),值是对应的服务器的名称（string）。这个映射用于根据哈希值查找对应的服务器。
}

func New(replicas int, fn HashFunc) *Map {
	m := &Map{
		hashFunc: fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// 这个函数用于添加实际节点到哈希环中
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// support hash tag
func getPartitionKey(key string) string {
	beg := strings.Index(key, "{")
	if beg == -1 {
		return key
	}
	end := strings.Index(key, "}")
	if end == -1 || end == beg+1 {
		return key
	}
	return key[beg+1 : end]
}

// 获取某个key实际对应的服务器
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	partitionKey := getPartitionKey(key)
	hash := int(m.hashFunc([]byte(partitionKey)))
	// Binary search
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
	if idx == len(m.keys) {
		idx = 0
	}
	return m.hashMap[m.keys[idx]]
}
