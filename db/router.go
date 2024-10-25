package db

func MakeRouter() map[string]CmdFunc {
	cmdMap := make(map[string]CmdFunc)
	cmdMap["ping"] = Ping
	cmdMap["get"] = Get
	cmdMap["set"] = Set
	cmdMap["del"] = Del
	cmdMap["mset"] = MSet
	cmdMap["mget"] = MGet
	cmdMap["getset"] = GetSet
	cmdMap["incr"] = Incr
	cmdMap["incrby"] = IncrBy
	cmdMap["incrbyfloat"] = IncrByFloat
	cmdMap["decr"] = Decr
	cmdMap["decrby"] = DecrBy
	cmdMap["decrbyfloat"] = DecrByFloat

	cmdMap["isexpired"] = IsExpired
	cmdMap["expire"] = Expire
	cmdMap["expireat"] = ExpireAt
	cmdMap["pexpire"] = PExpire
	cmdMap["pexpireat"] = PExpireAt
	cmdMap["ttl"] = TTL
	cmdMap["pttl"] = PTTL
	cmdMap["persist"] = Persist
	cmdMap["exists"] = Exists
	cmdMap["type"] = Type
	cmdMap["rename"] = Rename
	cmdMap["renamenx"] = RenameNX

	cmdMap["rpush"] = RPush
	cmdMap["lindex"] = LIndex
	cmdMap["llen"] = LLen
	cmdMap["lpop"] = LPop
	cmdMap["lpush"] = LPush
	cmdMap["lrange"] = LRange
	cmdMap["lrem"] = LRem
	cmdMap["lset"] = LSet
	cmdMap["rpop"] = RPop
	cmdMap["rpoplpush"] = RPopLPush

	cmdMap["hset"] = HSet
	cmdMap["hsetnx"] = HSetNX
	cmdMap["hget"] = HGet
	cmdMap["hexists"] = HExists
	cmdMap["hdel"] = HDel
	cmdMap["hlen"] = HLen
	cmdMap["hmset"] = HMSet
	cmdMap["hmget"] = HMGet
	cmdMap["hkeys"] = HKeys
	cmdMap["hvals"] = HVals
	cmdMap["hgetall"] = HGetAll
	cmdMap["hincrby"] = HIncrBy
	cmdMap["hincrbyfloat"] = HIncrByFloat

	cmdMap["sadd"] = SAdd
	cmdMap["sismember"] = SIsMember
	cmdMap["srem"] = SRem
	cmdMap["scard"] = SCard
	cmdMap["smembers"] = SMembers
	cmdMap["sinter"] = SInter
	cmdMap["sinterstore"] = SInterStore
	cmdMap["sunion"] = SUnion
	cmdMap["sunionstore"] = SUnionStore
	cmdMap["sdiff"] = SDiff
	cmdMap["sdiffstore"] = SDiffStore
	cmdMap["srandmember"] = SRandMember

	return cmdMap
}
