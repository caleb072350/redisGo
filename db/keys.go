package db

import (
	"redisGo/datastruct/list"
	"redisGo/interface/dict"
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"
	"time"
)

func Del(db *DB, args [][]byte) redis.Reply {
	if len(args) == 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'del' command")
	}
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	db.Locks(keys...)
	defer db.Unlocks(keys...)
	deleted := 0
	for _, key := range keys {
		_, exists := db.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return reply.MakeIntReply(int64(deleted))
}

func Exists(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'exists' command")
	}
	key := string(args[0])
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(1)
}

func FlushDB(db *DB, args [][]byte) redis.Reply {
	if len(args) != 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'flushdb' command")
	}
	db.Flush()
	return &reply.OkReply{}
}

func FlushAll(db *DB, args [][]byte) redis.Reply {
	if len(args) != 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'flushall' command")
	}
	db.Flush()
	return &reply.OkReply{}
}

func Type(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'type' command")
	}
	key := string(args[0])
	entity, exists := db.Get(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	case *list.LinkedList:
		return reply.MakeStatusReply("list")
	case dict.Dict:
		return reply.MakeStatusReply("hash")
	default:
		return &reply.UnknownErrReply{}
	}
}

func IsExpired(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'isexpired' command")
	}
	key := string(args[0])
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(-1)
	}
	expired := db.IsExpired(key)
	if expired {
		return reply.MakeIntReply(1)
	}
	return reply.MakeIntReply(0)
}

// 设置key多少秒之后过期
func Expire(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'expire' command")
	}
	key := string(args[0])
	ttlArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR invalid expire time in 'expire' command")
	}
	ttl := time.Duration(ttlArg) * time.Second
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	db.Expire(key, time.Now().Add(ttl))
	return reply.MakeIntReply(1)
}

func ExpireAt(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'expireat' command")
	}
	key := string(args[0])
	timestampArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR invalid expire time in 'expireat' command")
	}
	timestamp := time.Unix(timestampArg, 0)
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	db.Expire(key, timestamp)
	return reply.MakeIntReply(1)
}

// 功能与Expire相同，只不过单位为ms
func PExpire(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'pexpire' command")
	}
	key := string(args[0])
	ttlArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR invalid expire time in 'pexpire' command")
	}
	ttl := time.Duration(ttlArg) * time.Millisecond
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	db.Expire(key, time.Now().Add(ttl))
	return reply.MakeIntReply(1)
}

// 单位为ms
func PExpireAt(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'pexpireat' command")
	}
	key := string(args[0])
	timestampArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR invalid expire time in 'pexpireat' command")
	}
	timestamp := time.Unix(0, timestampArg*int64(time.Millisecond))
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	db.Expire(key, timestamp)
	return reply.MakeIntReply(1)
}

// 单位为秒
func TTL(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'ttl' command")
	}
	key := string(args[0])
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(-2)
	}
	raw, exists := db.TTLMap.Get(key)
	if !exists {
		return reply.MakeIntReply(-1)
	}
	expireTime, _ := raw.(time.Time)
	ttl := time.Until(expireTime)
	return reply.MakeIntReply(int64(ttl.Seconds()))
}

// 单位为毫秒
func PTTL(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'pttl' command")
	}
	key := string(args[0])
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(-2)
	}
	raw, exists := db.TTLMap.Get(key)
	if !exists {
		return reply.MakeIntReply(-1)
	}
	expireTime, _ := raw.(time.Time)
	ttl := time.Until(expireTime)
	return reply.MakeIntReply(int64(ttl.Milliseconds()))
}

func Persist(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'persist' command")
	}
	key := string(args[0])
	_, exists := db.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	_, exists = db.TTLMap.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	db.TTLMap.Remove(key)
	return reply.MakeIntReply(1)
}

func Rename(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	oldKey := string(args[0])
	newKey := string(args[1])
	db.Locks(oldKey, newKey)
	defer db.Unlocks(oldKey, newKey)
	entity, exists := db.Get(oldKey)
	if !exists {
		return reply.MakeErrReply("ERR no such key")
	}
	rawTTL, ok := db.TTLMap.Get(oldKey)
	db.Persist(oldKey)
	db.Persist(newKey)
	db.Put(newKey, entity)
	if ok {
		db.Expire(newKey, rawTTL.(time.Time))
	}
	return &reply.OkReply{}
}

func RenameNX(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'renamenx' command")
	}
	oldKey := string(args[0])
	newKey := string(args[1])
	db.Locks(oldKey, newKey)
	defer db.Unlocks(oldKey, newKey)
	_, exists := db.Get(newKey)
	if exists {
		return reply.MakeIntReply(0)
	}
	entity, ok := db.Get(oldKey)
	if !ok {
		return reply.MakeErrReply("ERR no such key")
	}
	db.Persist(oldKey)
	db.Persist(newKey)
	db.Put(newKey, entity)
	rawTTL, ok := db.TTLMap.Get(oldKey)
	if ok {
		db.Expire(newKey, rawTTL.(time.Time))
	}
	return reply.MakeIntReply(1)
}
