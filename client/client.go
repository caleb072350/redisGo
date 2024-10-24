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

	testCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n") // 设置 key-value

	// 测试 GET 命令
	testCommand(conn, "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n") // 获取 key 的值

	time.Sleep(time.Second)
}

func testCommand(conn net.Conn, command string) {
	fmt.Printf("发送命令:\n%s", command)

	// 发送命令
	fmt.Fprint(conn, command)
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}
	fmt.Println("响应:", response)
	if response[0] == '$' { // bulk response
		response, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取响应失败:", err)
			return
		}
		fmt.Println("响应:", response)
	}
}
