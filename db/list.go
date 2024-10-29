package db

import (
	List "redisGo/datastruct/list"
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"
)

func (db *DB) getAsList(key string) (*List.LinkedList, reply.ErrorReply) {
	entity, exists := db.Get(key)
	if !exists {
		return nil, nil
	}
	bytes, ok := entity.Data.(*List.LinkedList)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return bytes, nil
}

func (db *DB) getOrInitList(key string) (*List.LinkedList, bool, reply.ErrorReply) {
	list, errReply := db.getAsList(key)
	if errReply != nil {
		return nil, false, errReply
	}
	if list == nil {
		list = &List.LinkedList{}
		db.Put(key, &DataEntity{Data: list})
	}
	return list, true, nil
}

// 这个命令在list不存在的时候会新建一个list
func RPush(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rpush' command")
	}
	key := string(args[0])
	values := args[1:]

	db.Lock(key)
	defer db.Unlock(key)

	list, _, errReply := db.getOrInitList(key)
	if errReply != nil {
		return errReply
	}

	for _, value := range values {
		list.Add(value)
	}
	db.AddAof(makeAofCmd("rpush", args))
	return reply.MakeIntReply(int64(list.Len()))

}

func LIndex(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'lindex' command")
	}
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 32)
	if err != nil {
		return reply.MakeErrReply("ERR index is not integer")
	}
	index := int(index64)

	db.RLock(key)
	defer db.RUnlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return &reply.NullBulkReply{}
	}
	size := list.Len()
	if index < -1*size || index >= size {
		return &reply.NullBulkReply{}
	} else if index < 0 {
		index += size
	}
	return reply.MakeBulkReply(list.Get(index).([]byte))
}

func LLen(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'llen' command")
	}
	key := string(args[0])
	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(int64(list.Len()))
}

func LPop(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'lpop' command")
	}
	key := string(args[0])

	db.Lock(key)
	defer db.Unlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return &reply.NullBulkReply{}
	}
	val, _ := list.Remove(0).([]byte)
	if list.Len() == 0 {
		db.Data.Remove(key)
	}
	db.AddAof(makeAofCmd("lpop", args))
	return reply.MakeBulkReply(val)
}

func LPush(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR number of args for 'lpush' command")
	}
	key := string(args[0])
	values := args[1:]

	db.Lock(key)
	defer db.Unlock(key)

	list, _, errReply := db.getOrInitList(key)
	if errReply != nil {
		return errReply
	}
	for i := len(values) - 1; i >= 0; i-- {
		list.Insert(0, values[i])
	}
	db.AddAof(makeAofCmd("lpush", args))
	return reply.MakeIntReply(int64(list.Len()))
}

func LRange(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'lrange' command")
	}
	key := string(args[0])
	start64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR start is not integer")
	}
	start := int(start64)
	end64, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR end is not integer")
	}
	stop := int(end64)

	db.RLock(key)
	defer db.RUnlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return &reply.NullBulkReply{}
	}

	size := list.Len()
	if start < -1*size {
		start = 0
	} else if start < 0 {
		start = size + start
	} else if start >= size {
		return &reply.NullBulkReply{}
	}
	if stop < -1*size {
		stop = 0
	} else if stop < 0 {
		stop = size + stop + 1
	} else if stop < size {
		stop = stop + 1
	} else {
		stop = size
	}
	if start > stop {
		return &reply.NullBulkReply{}
	}
	slice := list.Range(start, stop)
	result := make([][]byte, len(slice))
	for i, v := range slice {
		result[i] = v.([]byte)
	}
	return reply.MakeMultiBulkReply(result)
}

func LRem(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'lrem' command")
	}
	key := string(args[0])
	count64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR count is not integer")
	}
	count := int(count64)
	val := args[2]

	db.Lock(key)
	defer db.Unlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return reply.MakeIntReply(0)
	}
	var removed int
	if count == 0 {
		removed = list.RemoveAllByVal(val)
	} else if count > 0 {
		removed = list.RemoveByVal(val, count)
	} else {
		removed = list.ReverseRemoveByVal(val, -count)
	}
	if list.Len() == 0 {
		db.Data.Remove(key)
	}
	db.AddAof(makeAofCmd("lrem", args))
	return reply.MakeIntReply(int64(removed))
}

func LSet(db *DB, args [][]byte) redis.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'lset' command")
	}
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR index is not integer")
	}
	index := int(index64)
	val := args[2]

	db.Lock(key)
	defer db.Unlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return reply.MakeErrReply("ERR list is nil")
	}
	size := list.Len()
	if index < -1*size {
		return reply.MakeErrReply("ERR index out of range")
	} else if index < 0 {
		index = index + size
	} else if index >= size {
		return reply.MakeErrReply("ERR index out of range")
	}
	list.Set(index, val)
	db.AddAof(makeAofCmd("lset", args))
	return &reply.OkReply{}
}

func RPop(db *DB, args [][]byte) redis.Reply {
	if len(args) != 1 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rpop' command")
	}
	key := string(args[0])

	db.Lock(key)
	defer db.Unlock(key)

	list, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if list == nil {
		return &reply.NullBulkReply{}
	}
	val, _ := list.RemoveLast().([]byte)
	if list.Len() == 0 {
		db.Data.Remove(key)
	}
	db.AddAof(makeAofCmd("rpop", args))
	return reply.MakeBulkReply(val)
}

func RPopLPush(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rpoplpush' command")
	}
	sourceKey := string(args[0])
	destKey := string(args[1])

	db.Locks(sourceKey, destKey)
	defer db.Unlocks(sourceKey, destKey)

	sourceList, errReply := db.getAsList(sourceKey)
	if errReply != nil {
		return errReply
	}
	if sourceList == nil {
		return reply.MakeErrReply("ERR source list is nil")
	}
	destList, errReply := db.getAsList(destKey)
	if errReply != nil {
		return errReply
	}
	if destList == nil {
		return reply.MakeErrReply("ERR dest list is nil")
	}
	val, _ := sourceList.RemoveLast().([]byte)
	destList.Insert(0, val)
	if sourceList.Len() == 0 {
		db.Remove(sourceKey)
	}
	db.AddAof(makeAofCmd("rpoplpush", args))
	return reply.MakeBulkReply(val)
}
