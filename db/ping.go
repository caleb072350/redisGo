package db

import (
	"redisGo/interface/redis"
	"redisGo/redis/reply"
)

type PongReply struct{}

func (p *PongReply) ToBytes() []byte {
	return []byte("+PONG\r\n")
}

func Ping(db *DB, args [][]byte) redis.Reply {
	if len(args) == 0 {
		return &PongReply{}
	} else if len(args) == 1 {
		return reply.MakeErrReply("\"" + string(args[0]) + "\"")
	} else {
		return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	}
}
