package main

import (
	"redisGo/lib/logger"
	"redisGo/server"
	"time"
)

func main() {
	settings := &logger.Settings{
		Path:       "logs",
		Name:       "Godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}
	logger.Setup(settings)

	cfg := &server.Config{
		Address:    ":6379",
		MaxConnect: 16,
		Timeout:    10 * time.Second,
	}
	server.ListenAndServe(cfg, &server.EchoServer{})
}
