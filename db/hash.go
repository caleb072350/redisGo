package db

import (
	Dict "redisGo/datastruct/dict"
	"redisGo/interface/dict"
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"

	"github.com/shopspring/decimal"
)

func (db *DB) getAsDict(key string) (dict.Dict, reply.ErrorReply) {
	entity, exists := db.Get(key)
	if !exists {
		return nil, nil
	}
	bytes, ok := entity.Data.(dict.Dict)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return bytes, nil
}

func (db *DB) getOrInitDict(key string) (dict dict.Dict, inited bool, errReply reply.ErrorReply) {
	dict, errReply = db.getAsDict(key)
	if errReply != nil {
		return nil, false, errReply
	}
	inited = false
	if dict == nil {
		dict = Dict.MakeSimple()
		db.Put(key, &DataEntity{Data: dict})
		inited = true
	}
	return dict, inited, nil
}

func HSet(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hset' command")
	}
	key := string(args[0])
	field := string(args[1])
	value := args[2]
	db.Lock(key)
	defer db.Unlock(key)
	dict, _, errReply := db.getOrInitDict(key)
	if errReply != nil {
		return errReply
	}
	res := dict.Put(field, value)
	db.addAof(makeAofCmd("hset", args))
	return reply.MakeIntReply(int64(res))
}

func HSetNX(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hsetnx' command")
	}
	key := string(args[0])
	field := string(args[1])
	value := args[2]
	db.Lock(key)
	defer db.Unlock(key)
	dict, _, errReply := db.getOrInitDict(key)
	if errReply != nil {
		return errReply
	}
	res := dict.PutIfAbsent(field, value)
	if res > 0 {
		db.addAof(makeAofCmd("hsetnx", args))
	}
	return reply.MakeIntReply(int64(res))
}

func HGet(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hget' command")
	}
	key := string(args[0])
	field := string(args[1])

	db.RLock(key)
	defer db.RUnlock(key)

	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return &reply.NullBulkReply{}
	}
	raw, exists := dict.Get(field)
	if !exists {
		return &reply.NullBulkReply{}
	}
	value, _ := raw.([]byte)
	return reply.MakeBulkReply(value)
}

func HExists(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hexists' command")
	}
	key := string(args[0])
	field := string(args[1])

	db.RLock(key)
	defer db.RUnlock(key)

	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return reply.MakeIntReply(0)
	}
	_, exists := dict.Get(field)
	if !exists {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(1)
}

func HDel(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hdel' command")
	}
	key := string(args[0])
	fields := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields[i-1] = string(args[i])
	}

	db.Lock(key)
	defer db.Unlock(key)

	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return reply.MakeIntReply(0)
	}
	count := 0
	for _, field := range fields {
		count += dict.Remove(field)
	}
	if dict.Len() == 0 {
		db.Remove(key)
	}
	if count > 0 {
		db.addAof(makeAofCmd("hdel", args))
	}
	return reply.MakeIntReply(int64(count))
}

func HLen(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hlen' command")
	}
	key := string(args[0])
	db.RLock(key)
	defer db.RUnlock(key)
	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(int64(dict.Len()))
}

func HMSet(db *DB, args [][]byte) redis.Reply {
	if len(args) < 3 || len(args)%2 == 0 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hmset' command")
	}
	key := string(args[0])
	size := (len(args) - 1) / 2
	fields := make([]string, size)
	values := make([][]byte, size)
	for i := 0; i < size; i++ {
		fields[i] = string(args[i*2+1])
		values[i] = args[i*2+2]
	}
	db.Lock(key)
	defer db.Unlock(key)
	dict, _, errReply := db.getOrInitDict(key)
	if errReply != nil {
		return errReply
	}
	for i, field := range fields {
		value := values[i]
		dict.Put(field, value)
	}
	db.addAof(makeAofCmd("hmset", args))
	return &reply.OkReply{}
}

func HMGet(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hmget' command")
	}
	key := string(args[0])
	fields := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields[i-1] = string(args[i])
	}
	db.RLock(key)
	defer db.RUnlock(key)

	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	result := make([][]byte, len(fields))
	for i, field := range fields {
		raw, exists := dict.Get(field)
		if !exists {
			result[i] = nil
		} else {
			value, _ := raw.([]byte)
			result[i] = value
		}
	}
	return reply.MakeMultiBulkReply(result)
}

func HKeys(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hkeys' command")
	}
	key := string(args[0])
	db.RLock(key)
	defer db.RUnlock(key)
	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	fields := make([][]byte, dict.Len())
	i := 0
	dict.ForEach(func(key string, _ interface{}) bool {
		fields[i] = []byte(key)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(fields)
}

func HVals(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hvals' command")
	}
	key := string(args[0])
	db.RLock(key)
	defer db.RUnlock(key)
	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	values := make([][]byte, dict.Len())
	i := 0
	dict.ForEach(func(_ string, value interface{}) bool {
		values[i] = value.([]byte)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(values)
}

func HGetAll(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hgetall' command")
	}
	key := string(args[0])
	db.RLock(key)
	defer db.RUnlock(key)
	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return errReply
	}
	if dict == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	result := make([][]byte, dict.Len()*2)
	i := 0
	dict.ForEach(func(key string, value interface{}) bool {
		result[i] = []byte(key)
		i++
		result[i], _ = value.([]byte)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(result[:i])
}

func HIncrBy(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hincrby' command")
	}
	key := string(args[0])
	field := string(args[1])
	delta, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	db.Lock(key)
	defer db.Unlock(key)
	dict, _, errReply := db.getOrInitDict(key)
	if errReply != nil {
		return errReply
	}
	value, exists := dict.Get(field)
	if exists {
		val, err := strconv.ParseInt(string(value.([]byte)), 10, 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not an integer or out of range")
		}
		val += delta
		bytes := []byte(strconv.FormatInt(val, 10))
		dict.Put(field, bytes)
		db.addAof(makeAofCmd("hincrby", args))
		return reply.MakeBulkReply(bytes)
	} else {
		dict.Put(field, args[2])
		db.addAof(makeAofCmd("hset", args))
		return reply.MakeBulkReply(args[2])
	}
}

func HIncrByFloat(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'hincrbyfloat' command")
	}
	key := string(args[0])
	field := string(args[1])
	delta, err := decimal.NewFromString(string(args[2]))
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	db.Lock(key)
	defer db.Unlock(key)
	dict, _, errReply := db.getOrInitDict(key)
	if errReply != nil {
		return errReply
	}
	value, exists := dict.Get(field)
	if exists {
		val, err := decimal.NewFromString(string(value.([]byte)))
		if err != nil {
			return reply.MakeErrReply("ERR value is not a valid float")
		}
		result := val.Add(delta)
		dict.Put(field, []byte(result.String()))
		db.addAof(makeAofCmd("hincrbyfloat", args))
		return reply.MakeBulkReply([]byte(result.String()))
	} else {
		dict.Put(field, args[2])
		db.addAof(makeAofCmd("hset", args))
		return reply.MakeBulkReply(args[2])
	}
}
