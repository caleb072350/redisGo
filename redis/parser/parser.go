package parser

import (
	"io"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"runtime/debug"
)

type Payload struct {
	Data redis.Reply
	Err  error
}

func Parse(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(debug.Stack())
			}
		}()
		parse0(reader, ch)
	}()
	return ch
}

type readState struct {
	downloading       bool
	expectedArgsCount int
	receivedCount     int
	msgType           byte
	args              [][]byte
	fixedLen          int64
}

func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && s.receivedCount == s.expectedArgsCount
}

func parse0(reader io.Reader, ch chan<- *Payload) {

}
