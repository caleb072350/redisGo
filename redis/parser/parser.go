package parser

import (
	"bufio"
	"errors"
	"io"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"runtime/debug"
	"strconv"
	"strings"
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
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte

	for {
		// read line
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				ch <- &Payload{Err: err}
				close(ch)
				return
			} else {
				ch <- &Payload{Err: err}
				state = readState{}
				continue
			}
		}

		// parse line
		if !state.downloading {
			// receive new response
			if msg[0] == '*' {
				// multi bulk response
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.fixedLen == -1 {
					ch <- &Payload{Data: &reply.NullBulkReply{}}
					state = readState{}
					continue
				}
			} else {
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
				continue
			}
		} else {
			err = readBulkBody(msg, &state)
			if err != nil {
				ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
				state = readState{}
				continue
			}
			if state.finished() {
				var result redis.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
			}
		}
	}
}

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.fixedLen == 0 { // read normal line
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		msg = make([]byte, state.fixedLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.fixedLen = 0
	}
	return msg, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.downloading = true
		state.expectedArgsCount = int(expectedLine)
		state.receivedCount = 0
		state.msgType = msg[0]
		state.args = make([][]byte, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.fixedLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.fixedLen == -1 {
		return nil
	} else if state.fixedLen > 0 {
		state.msgType = msg[0]
		state.downloading = true
		state.expectedArgsCount = 1
		state.receivedCount = 0
		state.args = make([][]byte, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (redis.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result redis.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		v, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(v)
	default:
		strs := strings.Split(str, " ")
		args := make([][]byte, len(strs))
		for i, s := range strs {
			args[i] = []byte(s)
		}
		result = reply.MakeMultiBulkReply(args)
	}
	return result, nil
}

func readBulkBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		state.fixedLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.fixedLen <= 0 {
			state.args[state.receivedCount] = []byte{}
			state.receivedCount++
			state.fixedLen = 0
		}
	} else {
		state.args[state.receivedCount] = line
		state.receivedCount++
	}
	return nil
}
