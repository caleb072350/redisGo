package client

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/lib/sync/wait"
	"redisGo/redis/reply"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn        net.Conn
	sendingReqs chan *Request
	waitingReqs chan *Request
	ticker      *time.Ticker
	addr        string

	ctx        context.Context
	cancelFunc context.CancelFunc
	writing    *sync.WaitGroup
}

type Request struct {
	args      [][]byte
	reply     redis.Reply
	heartbeat bool
	waiting   *wait.Wait
	err       error
}

const (
	chanSize = 256
	maxWait  = 3 * time.Second
)

func MakeClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		conn:        conn,
		addr:        addr,
		sendingReqs: make(chan *Request, chanSize),
		waitingReqs: make(chan *Request, chanSize),
		ctx:         ctx,
		cancelFunc:  cancel,
		writing:     &sync.WaitGroup{},
	}, nil
}

func (client *Client) Start() {
	client.ticker = time.NewTicker(10 * time.Second)
	go client.handleWrite()
	go func() {
		err := client.handleRead()
		logger.Info(err)
	}()
	go client.heartbeat()
}

func (client *Client) Close() {
	close(client.sendingReqs)
	client.writing.Wait()
	client.cancelFunc()
	_ = client.conn.Close()
	close(client.waitingReqs)
}

func (client *Client) Send(args [][]byte) redis.Reply {
	request := &Request{
		args:      args,
		heartbeat: false,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.sendingReqs <- request
	timeout := request.waiting.WaitWithTimeout(maxWait)
	if timeout {
		return reply.MakeErrReply("server time out")
	}
	if request.err != nil {
		return reply.MakeErrReply("request failed")
	}
	return request.reply
}

func (client *Client) heartbeat() {
loop:
	for {
		select {
		case <-client.ticker.C:
			client.sendingReqs <- &Request{
				args:      [][]byte{[]byte("PING")},
				heartbeat: true,
			}
		case <-client.ctx.Done():
			break loop
		}
	}
}

func (client *Client) handleWrite() {
loop:
	for {
		select {
		case req := <-client.sendingReqs:
			client.writing.Add(1)
			client.sendRequest(req)
		case <-client.ctx.Done():
			break loop
		}
	}
}

func (client *Client) handleConnectionError(_ error) error {
	err1 := client.conn.Close()
	if err1 != nil {
		if opErr, ok := err1.(*net.OpError); ok {
			if opErr.Err.Error() != "use of closed network connection" {
				return err1
			}
		} else {
			return err1
		}
	}
	conn, err1 := net.Dial("tcp", client.addr)
	if err1 != nil {
		logger.Error(err1)
		return err1
	}
	client.conn = conn
	go func() {
		_ = client.handleRead()
	}()
	return nil
}

func (client *Client) sendRequest(req *Request) {
	bytes := reply.MakeMultiBulkReply(req.args).ToBytes()
	_, err := client.conn.Write(bytes)
	i := 0
	for err != nil && i < 3 {
		err = client.handleConnectionError(err)
		if err == nil {
			_, err = client.conn.Write(bytes)
		}
		i++
	}
	if err == nil {
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
		client.writing.Done()
	}
}

func (client *Client) finishRequest(reply redis.Reply) {
	request := <-client.waitingReqs
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
}

func (client *Client) handleRead() error {
	reader := bufio.NewReader(client.conn)
	downloading := false
	expectedArgsCount := 0
	receivedCount := 0
	msgType := byte(0)
	var args [][]byte
	var fixedLen int64 = 0
	var err error
	var msg []byte
	for {
		if fixedLen == 0 { // read normal line
			msg, err = reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					logger.Info("connection close")
				} else {
					logger.Warn(err)
				}
				return errors.New("connection error")
			}
			if len(msg) == 0 || msg[len(msg)-2] != '\r' {
				return errors.New("protocol error: " + string(msg))
			}
		} else { // read bulk line (binary safe)
			msg = make([]byte, fixedLen+2)
			_, err = io.ReadFull(reader, msg)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return errors.New("connection close")
				} else {
					return err
				}
			}
			if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
				return errors.New("protocol error: " + string(msg))
			}
			fixedLen = 0
		}

		// parse line
		if !downloading {
			// receive new response
			if msg[0] == '*' { // multibulk response
				expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					return errors.New("protocol error: " + string(msg))
				}
				if expectedLine == 0 {
					client.finishRequest(&reply.EmptyMultiBulkReply{})
				} else if expectedLine > 0 {
					msgType = msg[0]
					downloading = true
					expectedArgsCount = int(expectedLine)
					receivedCount = 0
					args = make([][]byte, expectedLine)
				} else {
					return errors.New("protocol error: " + string(msg))
				}
			} else if msg[0] == '$' { // bulk response
				fixedLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
				if err != nil {
					return err
				}
				if fixedLen == -1 {
					client.finishRequest(&reply.NullBulkReply{})
					fixedLen = 0
				} else if fixedLen > 0 {
					msgType = msg[0]
					downloading = true
					args = make([][]byte, 1)
					expectedArgsCount = 1
					receivedCount = 0
				} else {
					return errors.New("protocol error: " + string(msg))
				}
			} else { // single line response
				str := strings.TrimSuffix(string(msg), "\n")
				str = strings.TrimSuffix(str, "\r")
				var result redis.Reply
				switch msg[0] {
				case '+':
					result = reply.MakeStatusReply(str[1:])
				case '-':
					result = reply.MakeErrReply(str[1:])
				case ':':
					val, err := strconv.ParseInt(str[1:], 10, 64)
					if err != nil {
						return errors.New("protocol error: " + string(msg))
					}
					result = reply.MakeIntReply(val)
				}
				client.finishRequest(result)
				fixedLen = 0
			}
		} else { // reveiving args
			line := msg[0 : len(msg)-2]
			if line[0] == '$' {
				fixedLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
				if err != nil {
					return errors.New("protocol error: " + string(msg))
				}
				if fixedLen <= 0 {
					args[receivedCount] = []byte{}
					receivedCount++
					fixedLen = 0
				}
			} else {
				args[receivedCount] = line
				receivedCount++
			}

			// if sending finished
			if receivedCount == expectedArgsCount {
				downloading = false

				if msgType == '*' {
					reply := reply.MakeMultiBulkReply(args)
					client.finishRequest(reply)
				} else if msgType == '$' {
					reply := reply.MakeBulkReply(args[0])
					client.finishRequest(reply)
				}
				// finish reply
				expectedArgsCount = 0
				receivedCount = 0
				args = nil
				msgType = byte(0)
			}
		}
	}
}
