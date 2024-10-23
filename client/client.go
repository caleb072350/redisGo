package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	// 发送命令
	fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")

	// 读取并打印响应
	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)
	for {
		// line, err := reader.ReadString('\n')
		content, err := reader.Read(buf)
		if err != nil || content == 0 {
			fmt.Println("read over")
			break
		}
		line := string(buf)
		fmt.Println(line)
	}

	time.Sleep(time.Second)
}
