package dict

type Consumer func(key string, val interface{}) bool

type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	Put(key string, val interface{}) int
	PutIfExists(key string, val interface{}) int
	PutIfAbsent(key string, val interface{}) int
	Remove(key string) int
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(n int) []string
	RandomDistinctKeys(n int) []string
}
