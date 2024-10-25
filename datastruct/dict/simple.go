package dict

import "redisGo/interface/dict"

type SimpleDict struct {
	m map[string]interface{}
}

func MakeSimple() *SimpleDict {
	return &SimpleDict{m: make(map[string]interface{})}
}

func (d *SimpleDict) Get(key string) (val interface{}, exists bool) {
	val, exists = d.m[key]
	return
}

func (d *SimpleDict) Len() int {
	return len(d.m)
}

func (d *SimpleDict) Put(key string, val interface{}) int {
	_, exists := d.m[key]
	d.m[key] = val
	if !exists {
		return 1
	} else {
		return 0
	}
}

func (d *SimpleDict) PutIfExists(key string, val interface{}) int {
	_, exists := d.m[key]
	if exists {
		d.m[key] = val
		return 1
	} else {
		return 0
	}
}

func (d *SimpleDict) PutIfAbsent(key string, val interface{}) int {
	_, exists := d.m[key]
	if exists {
		return 0
	} else {
		d.m[key] = val
		return 1
	}
}

func (d *SimpleDict) Remove(key string) int {
	_, exists := d.m[key]
	if exists {
		delete(d.m, key)
		return 1
	} else {
		return 0
	}
}

func (d *SimpleDict) ForEach(consumer dict.Consumer) {
	for k, v := range d.m {
		if !consumer(k, v) {
			break
		}
	}
}

func (d *SimpleDict) Keys() []string {
	result := make([]string, len(d.m))
	i := 0
	for k := range d.m {
		result[i] = k
		i++
	}
	return result
}

func (d *SimpleDict) RandomKeys(n int) []string {
	result := make([]string, n)
	for i := 0; i < n; i++ {
		for key := range d.m {
			result[i] = key
			break
		}
	}
	return result
}

func (d *SimpleDict) RandomDistinctKeys(n int) []string {
	size := n
	if size > len(d.m) {
		size = len(d.m)
	}
	result := make([]string, size)
	i := 0
	for key := range d.m {
		result[i] = key
		i++
		if i == size {
			break
		}
	}
	return result
}
