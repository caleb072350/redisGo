package server

import (
	"context"
	"io"
	"net"
	"redisGo/db"
	"redisGo/lib/logger"
	"redisGo/lib/sync/atomic"
	"redisGo/redis/parser"
	"redisGo/redis/reply"
	"strings"
	"sync"
)

var (
	UnknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

/*
 * 一个实现了redis协议的TCP Handler
 */

type RedisHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         db.DB
	closing    atomic.AtomicBool
}

func MakeRedisHandler() *RedisHandler {
	return &RedisHandler{
		db: *db.MakeDB(),
	}
}

func (s *RedisHandler) closeClient(client *Client) {
	_ = client.Close()
	s.db.AfterClientClose(client)
	s.activeConn.Delete(client)
}

func (h *RedisHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		_ = conn.Close()
	}
	client := MakeClient(conn)
	h.activeConn.Store(client, 1)

	ch := parser.Parse(conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF || payload.Err == io.ErrUnexpectedEOF || strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(client)
				logger.Info("connection closed: " + client.conn.RemoteAddr().String())
				return
			}
		} else {
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.conn.RemoteAddr().String())
				return
			}
			continue
		}

		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(UnknownErrReplyBytes)
		}
	}
}

func (s *RedisHandler) Close() error {
	logger.Info("redis handler shuting down...")
	s.closing.Set(true)
	s.activeConn.Range(func(key, _ any) bool {
		client := key.(*Client)
		_ = client.Close()
		return true
	})
	s.db.Close()
	return nil
}
