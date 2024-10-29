package cluster

import "redisGo/interface/redis"

func defaultFunc(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply {
	key := string(args[1])
	return cluster.Relay(key, c, args)
}

func MakeRouter() map[string]CmdFunc {
	router := make(map[string]CmdFunc)

	router["del"] = Del
	return router
}
