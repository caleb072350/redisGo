package db

import (
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	upsertPolicy = iota // default
	insertPolicy        // set nx
	updatePolicy        // set ex
)

func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
	entity, ok := db.Get(key)
	if !ok {
		return nil, reply.MakeErrorReply("ERR key not found")
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
	if ttl != unlimitedTTL {
		expireTime := time.Now().Add(time.Duration(ttl) * time.Millisecond)
		db.Expire(key, expireTime)
	} else {
		db.Persist(key)
	}
	return &reply.OkReply{}
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

func GetSet(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'getset' command")
	}
	key := string(args[0])
	value := args[1]
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	db.Data.PutIfExists(key, &DataEntity{Data: value})
	return reply.MakeBulkReply(bytes)
}

func Incr(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'incr' command")
	}
	key := string(args[0])
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Data.PutIfExists(key, &DataEntity{Data: []byte(strconv.FormatInt(i+1, 10))})
	return reply.MakeIntReply(i + 1)
}

func IncrBy(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'incrby' command")
	}
	key := string(args[0])
	raw := string(args[1])
	delta, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Data.PutIfExists(key, &DataEntity{Data: []byte(strconv.FormatInt(i+delta, 10))})
	return reply.MakeIntReply(i + delta)
}

func IncrByFloat(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'incrbyfloat' command")
	}
	key := string(args[0])
	raw := string(args[1])
	delta, err := decimal.NewFromString(raw)
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := decimal.NewFromString(string(bytes))
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	resultBytes := []byte(i.Add(delta).String())
	db.Data.PutIfExists(key, &DataEntity{Data: resultBytes})
	return reply.MakeBulkReply(resultBytes)
}

func Decr(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'decr' command")
	}
	key := string(args[0])
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Data.PutIfExists(key, &DataEntity{Data: []byte(strconv.FormatInt(i-1, 10))})
	return reply.MakeIntReply(i - 1)
}

func DecrBy(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'decrby' command")
	}
	key := string(args[0])
	raw := string(args[1])
	delta, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Data.PutIfExists(key, &DataEntity{Data: []byte(strconv.FormatInt(i-delta, 10))})
	return reply.MakeIntReply(i - delta)
}

func DecrByFloat(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'decrbyfloat' command")
	}
	key := string(args[0])
	raw := string(args[1])
	delta, err := decimal.NewFromString(raw)
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	db.Lock(key)
	defer db.Unlock(key)
	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes == nil {
		bytes = []byte("0")
	}
	i, err := decimal.NewFromString(string(bytes))
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	resultBytes := []byte(i.Sub(delta).String())
	db.Data.PutIfExists(key, &DataEntity{Data: resultBytes})
	return reply.MakeBulkReply(resultBytes)
}
