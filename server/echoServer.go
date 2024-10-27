package server

import (
	"bufio"
	"context"
	"io"
	"net"
	"redisGo/lib/logger"
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
	if s.closing.Get() {
		// closing handler refuse new connection
		conn.Close()
	}
	client := &Client{
		Conn: conn,
	}
	s.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				s.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		conn.Write(b)
		client.Waiting.Done()
	}
}

func (s *EchoServer) Close() error {
	logger.Info("handler shuting down...")
	s.closing.Set(true)
	s.activeConn.Range(func(key interface{}, _ interface{}) bool {
		client := key.(*Client)
		_ = client.Conn.Close()
		return true
	})
	return nil
}
