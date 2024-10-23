package server

import (
	"context"
	"net"
	"redisGo/lib/sync/atomic"
	"redisGo/lib/sync/wait"
	"sync"
)

type EchoServer struct {
	activeConn sync.Map
	closing    atomic.AtomicBool
}

func MakeEchoServer() *EchoServer {
	return &EchoServer{}
}

type Client struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (s *EchoServer) Handle(ctx context.Context, conn net.Conn) {
}

func (s *EchoServer) Close() error {
	return nil
}
