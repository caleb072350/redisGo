package pubsub

import (
	Dict "redisGo/datastruct/dict"
	"redisGo/datastruct/lock"
	"redisGo/interface/dict"
)

type Hub struct {
	// channel -> list(*Client)
	subs dict.Dict
	// lock channel
	subsLocker *lock.LockMap
}

func MakeHub() *Hub {
	return &Hub{
		subs:       Dict.MakeConcurrent(4),
		subsLocker: lock.Make(16),
	}
}
