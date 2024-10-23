package tcp

import (
	"net"
	"redisGo/lib/sync/atomic"
	"redisGo/lib/sync/wait"
	"time"
)

type Client struct {
	conn net.Conn

	// 带超时的wait
	waitingReply wait.Wait

	// 是否在发送请求过程中
	sending atomic.AtomicBool

	// bulk msg lineCount - 1
	expectedLineCount uint32

	// sent line count, exclude first line
	sentLineCount uint32

	// sent lines, exclude first line
	sentLines [][]byte
}

func (c *Client) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	c.conn.Close()
	return nil
}

func MakeClient(conn net.Conn) *Client {
	return &Client{
		conn: conn,
	}
}
