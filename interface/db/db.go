package db

import "redisGo/interface/redis"

type DB interface {
	Exec([][]byte) redis.Reply
}
