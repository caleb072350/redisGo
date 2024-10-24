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

func Get(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'get' command")
	}
	key := string(args[0])
	val, ok := db.Data.Get(key)
	if !ok {
		return &reply.NullBulkReply{}
	}
	entity, _ := val.(*DataEntity)
	if entity.Code == StringCode {
		bytes, ok := entity.Data.([]byte)
		if !ok {
			return &reply.UnknownErrReply{}
		}
		return reply.MakeBulkReply(bytes)
	} else {
		return reply.MakeErrReply("ERR get only support string")
	}
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
		Code: StringCode,
		TTL:  ttl,
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
