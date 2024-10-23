package tcp

import (
	"bufio"
	"context"
	"net"
	"redisGo/db"
	"redisGo/lib/logger"
	"redisGo/lib/sync/atomic"
	"redisGo/redis/parser"
	"strconv"
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

func (s *RedisHandler) Handle(ctx context.Context, conn net.Conn) {
	if s.closing.Get() {
		conn.Close()
	}
	client := &Client{
		conn: conn,
	}
	s.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			logger.Warn(err)
			client.Close()
			s.activeConn.Delete(client)
			return
		}
		if len(msg) == 0 {
			continue
		}

		if !client.sending.Get() {
			if msg[0] == '*' {
				expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					client.conn.Write(UnknownErrReplyBytes)
					continue
				}
				expectedLine *= 2
				client.waitingReply.Add(1)
				client.sending.Set(true)
				client.expectedLineCount = uint32(expectedLine)
				client.sentLineCount = 0
				client.sentLines = make([][]byte, expectedLine)
			} else {
				// TODO: text protocol
			}
		} else {
			// receiving following part of a request
			client.sentLines[client.sentLineCount] = msg[0 : len(msg)-2]
			client.sentLineCount++

			if client.sentLineCount == client.expectedLineCount {
				client.sending.Set(false)

				if len(client.sentLines)%2 != 0 {
					client.conn.Write(UnknownErrReplyBytes)
					client.expectedLineCount = 0
					client.sentLineCount = 0
					client.sentLines = nil
					client.waitingReply.Done()
					continue
				}

				// send reply
				args := parser.Parse(client.sentLines)
				result := s.db.Exec(args) // 这里是执行redis命令的入口
				if result != nil {
					conn.Write(result.ToBytes())
				} else {
					conn.Write(UnknownErrReplyBytes)
				}

				// finish reply
				client.expectedLineCount = 0
				client.sentLineCount = 0
				client.sentLines = nil
				client.waitingReply.Done()
			}
		}

	}
}

func (s *RedisHandler) Close() error {
	logger.Info("redis handler shuting down...")
	s.closing.Set(true)
	s.activeConn.Range(func(key, value any) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}
