package db

import "redisGo/interface/redis"

type DB interface {
	Exec(client redis.Client, args [][]byte) redis.Reply
	AfterClientClose(c redis.Client)
}
