package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"redisGo/db"
	"redisGo/lib/logger"
	"redisGo/lib/sync/atomic"
	"redisGo/redis/reply"
	"strconv"
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

func (s *RedisHandler) Handle(ctx context.Context, conn net.Conn) {
	if s.closing.Get() {
		conn.Close()
	}
	client := &Client{
		conn: conn,
	}
	s.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	var fixedLen int64 = 0
	var err error
	var msg []byte
	for {
		if fixedLen == 0 {
			msg, err = reader.ReadBytes('\n')
		} else {
			msg = make([]byte, fixedLen+2)
			_, err = io.ReadFull(reader, msg)
			fixedLen = 0
		}
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				logger.Info("connection close")
			} else {
				logger.Warn(err)
			}
			client.Close()
			s.activeConn.Delete(client)
			return
		}
		if len(msg) == 0 {
			continue
		}

		if !client.uploading.Get() {
			if msg[0] == '*' {
				// bulk multi msg
				expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					client.conn.Write(UnknownErrReplyBytes)
					continue
				}

				client.waitingReply.Add(1)
				client.uploading.Set(true)
				client.expectedArgsCount = uint32(expectedLine)
				client.receivedCount = 0
				client.args = make([][]byte, expectedLine)
			} else {
				// TODO: text protocol
				// remove \r or \n or \r\n in the end of line
				str := strings.TrimSuffix(string(msg), "\n")
				str = strings.TrimSuffix(str, "\r")
				strs := strings.Split(str, " ")
				args := make([][]byte, len(strs))
				for i, s := range strs {
					args[i] = []byte(s)
				}
				result := s.db.Exec(client, args)
				if result != nil {
					_ = client.Write(result.ToBytes())
				} else {
					_ = client.Write(UnknownErrReplyBytes)
				}
			}
		} else {
			// receiving following part of a request
			line := msg[0 : len(msg)-2]
			if line[0] == '$' {
				fixedLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
				if err != nil {
					errReply := &reply.ProtocolErrReply{Msg: err.Error()}
					_, _ = client.conn.Write(errReply.ToBytes())
				}
				if fixedLen <= 0 {
					errReply := reply.ProtocolErrReply{Msg: "invalid multibulk length"}
					_, _ = client.conn.Write(errReply.ToBytes())
				}
			} else {
				client.args[client.receivedCount] = line
				client.receivedCount++
			}

			if client.receivedCount == client.expectedArgsCount {
				client.uploading.Set(false)

				// send reply
				result := s.db.Exec(client, client.args) // 这里是执行redis命令的入口
				if result != nil {
					_, _ = conn.Write(result.ToBytes())
				} else {
					_, _ = conn.Write(UnknownErrReplyBytes)
				}

				// finish reply
				client.expectedArgsCount = 0
				client.receivedCount = 0
				client.args = nil
				client.waitingReply.Done()
			}
		}

	}
}

func (s *RedisHandler) Close() error {
	logger.Info("redis handler shuting down...")
	s.closing.Set(true)
	s.activeConn.Range(func(key, _ any) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}
