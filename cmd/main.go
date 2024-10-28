package main

import (
	"fmt"
	"redisGo/config"
	"redisGo/lib/logger"
	RedisServer "redisGo/redis/server"
	"redisGo/tcp"
	"time"
)

func main() {
	config.SetupConfig("redis.conf")
	settings := &logger.Settings{
		Path:       "logs",
		Name:       "Godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}
	logger.Setup(settings)

	cfg := &tcp.Config{
		Address:    fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		MaxConnect: uint32(config.Properties.MaxClients),
		Timeout:    2 * time.Second,
	}
	// server.ListenAndServe(cfg, &server.EchoServer{})
	handler := RedisServer.MakeRedisHandler()
	tcp.ListenAndServe(cfg, handler)
}
