package db

import (
	"fmt"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"runtime/debug"
	"strings"
)

type CmdFunc func(args [][]byte) redis.Reply

type DB struct {
	cmdMap map[string]CmdFunc
}

type UnknownErrReply struct{}

func (r *UnknownErrReply) ToBytes() []byte {
	return []byte("-ERR unknown command\r\n")
}

func (db *DB) Exec(args [][]byte) (result redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &UnknownErrReply{}
		}
	}()

	cmd := strings.ToLower(string(args[0]))
	CmdFunc, ok := db.cmdMap[cmd]
	if !ok {
		return reply.MakeErrReply("ERR unknown command `" + cmd + "`")
	}
	if len(args) > 1 {
		result = CmdFunc(args[1:])
	} else {
		result = CmdFunc([][]byte{})
	}
	return
}

func MakeDB() *DB {
	cmdMap := make(map[string]CmdFunc)
	cmdMap["ping"] = Ping

	return &DB{
		cmdMap: cmdMap,
	}
}
