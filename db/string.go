package db

import (
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"
	"strings"
)

const (
	upsertPolicy = iota // default
	insertPolicy        // set nx
	updatePolicy        // set ex
)

func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
	entity, ok := db.Get(key)
	if !ok {
		return nil, nil
	}
	bytes, ok := entity.Data.([]byte)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return bytes, nil
}

func Get(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'get' command")
	}
	key := string(args[0])
	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bytes == nil {
		return &reply.NullBulkReply{}
	}
	return reply.MakeBulkReply(bytes)
}

const unlimitedTTL int64 = 0

func Set(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'set' command")
	}

	key := string(args[0])
	value := args[1]
	policy := upsertPolicy
	ttl := unlimitedTTL

	if len(args) > 2 {
		for i := 2; i < len(args); i++ {
			arg := strings.ToUpper(string(args[i]))
			if arg == "NX" {
				if policy == updatePolicy {
					return &reply.SyntaxErrReply{}
				}
				policy = insertPolicy
			} else if arg == "XX" {
				if policy == insertPolicy {
					return &reply.SyntaxErrReply{}
				}
				policy = updatePolicy
			} else if arg == "EX" {
				if ttl != unlimitedTTL {
					return &reply.SyntaxErrReply{}
				}
				if i+1 > len(args) {
					return &reply.SyntaxErrReply{}
				}
				ttlArg, err := strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					return &reply.SyntaxErrReply{}
				}
				if ttlArg < 0 {
					return &reply.SyntaxErrReply{}
				}
				ttl = ttlArg * 1000
				i++
			} else if arg == "PX" {
				if ttl != unlimitedTTL {
					return &reply.SyntaxErrReply{}
				}
				if i+1 > len(args) {
					return &reply.SyntaxErrReply{}
				}
				ttlArg, err := strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					return &reply.SyntaxErrReply{}
				}
				if ttlArg < 0 {
					return &reply.SyntaxErrReply{}
				}
				ttl = ttlArg
				i++
			}
		}
	}

	entity := &DataEntity{
		Data: value,
	}
	switch policy {
	case upsertPolicy:
		db.Data.Put(key, entity)
	case insertPolicy:
		db.Data.PutIfAbsent(key, entity)
	case updatePolicy:
		db.Data.PutIfExists(key, entity)
	}
	return &reply.OkReply{}
}

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

func MSet(db *DB, args [][]byte) redis.Reply {
	if len(args)%2 != 0 || len(args) == 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'mset' command")
	}
	keys := make([]string, len(args)/2)
	for i := 0; i < len(args)/2; i++ {
		keys[i] = string(args[i*2])
	}
	db.Locks(keys...)
	defer db.Unlocks(keys...)

	for i := 0; i < len(args); {
		key := string(args[i])
		value := args[i+1]
		entity := &DataEntity{
			Data: value,
		}
		db.Data.Put(key, entity)
		i += 2
	}
	return &reply.OkReply{}
}

func MGet(db *DB, args [][]byte) redis.Reply {
	if len(args) == 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'mget' command")
	}
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	db.RLocks(keys...)
	defer db.RUnlocks(keys...)
	values := make([][]byte, len(args))
	for i, key := range keys {
		entity, exists := db.Get(key)
		if !exists {
			values[i] = nil
			continue
		}
		values[i] = entity.Data.([]byte)
	}
	return reply.MakeMultiBulkReply(values)
}
