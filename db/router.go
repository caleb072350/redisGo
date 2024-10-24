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
	return cmdMap
}
