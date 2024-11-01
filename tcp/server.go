package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"redisGo/interface/tcp"
	"redisGo/lib/logger"
	"redisGo/lib/sync/atomic"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"maxConnect"`
	Timeout    time.Duration `yaml:"timeout"`
}

func ListenAndServe(cfg *Config, handler tcp.Handler) {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		logger.Fatal(fmt.Sprintf("listen err: %v", err))
	}

	// listen signal
	var closing atomic.AtomicBool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP:
			logger.Info(fmt.Sprintf("get signal: %s", sig.String()))
			logger.Info("server is shuting down...")
			closing.Set(true)
			listener.Close() // listener.Accept will return err immediately
		}
	}()

	// listen port
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	// closing listener than closing handler while shuting down
	defer handler.Close()
	defer listener.Close()

	ctx, _ := context.WithCancel(context.Background())
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if closing.Get() {
				return
			}
			logger.Error(fmt.Sprintf("accept err: %v", err))
			continue
		}
		logger.Info(fmt.Sprintf("accept new connection: %s", conn.RemoteAddr().String()))
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
}
