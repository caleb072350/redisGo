package set

import (
	Dict "redisGo/datastruct/dict"
	"redisGo/interface/dict"
)

type Set struct {
	dict dict.Dict
}

func Make(hint int) *Set {
	return &Set{
		dict: Dict.MakeConcurrent(hint),
	}
}

func MakeFromVals(members ...string) *Set {
	set := &Set{
		dict: Dict.MakeConcurrent(10),
	}
	for _, member := range members {
		set.Add(member)
	}
	return set
}

func (set *Set) Add(val string) int {
	return set.dict.Put(val, true)
}

func (set *Set) Remove(val string) int {
	return set.dict.Remove(val)
}

func (set *Set) Has(val string) bool {
	_, exists := set.dict.Get(val)
	return exists
}

func (set *Set) Len() int {
	return set.dict.Len()
}

func (set *Set) ToSlice() []string {
	slice := make([]string, set.Len())
	i := 0
	set.dict.ForEach(func(key string, _ interface{}) bool {
		if i < len(slice) {
			slice[i] = key
		} else {
			slice = append(slice, key)
		}
		i++
		return true
	})
	return slice
}

func (set *Set) ForEach(consumer func(key string) bool) {
	set.dict.ForEach(func(key string, _ interface{}) bool {
		return consumer(key)
	})
}

func (set *Set) Intersect(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}
	result := Make(0)
	another.ForEach(func(member string) bool {
		if set.Has(member) {
			result.Add(member)
		}
		return true
	})
	return result
}

func (set *Set) Union(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}
	result := Make(0)
	another.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	set.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	return result
}

func (set *Set) Diff(another *Set) *Set {
	if set == nil {
		panic("set is nil")
	}
	result := Make(0)
	set.ForEach(func(member string) bool {
		if !another.Has(member) {
			result.Add(member)
		}
		return true
	})
	return result
}

func (set *Set) RandomMembers(limit int) []string {
	return set.dict.RandomKeys(limit)
}

func (set *Set) RandomDistinctMembers(limit int) []string {
	return set.dict.RandomDistinctKeys(limit)
}
