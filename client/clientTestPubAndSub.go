package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	// testCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	// 测试 订阅
	testCommand(conn, "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n")
	//fmt.Fprintln(conn, "*2\r\n$9\r\nSUBSCRIBE\r\n$11\r\ntestchannel\r\n")

	// go testReadChannel(conn)

	// 保持主goroutine 运行，否则程序会退出
	// select {}
	// testCommand(conn, "*5\r\n$5\r\nRPUSH\r\n$3\r\narr\r\n$1\r\n1\r\n$1\r\n2\r\n$1\r\n3\r\n")
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
	} else if response[0] == '*' {
		cnt, err := strconv.ParseInt(response[1:len(response)-2], 10, 32)
		if err != nil {
			fmt.Println("解析multibulk协议失败:", err)
			return
		}
		for i := 0; i < int(cnt*2); i++ {
			response, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("读取响应失败:", err)
				return
			}
			fmt.Println("响应:", response)
		}
	}

}

func testReadChannel(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取响应失败: ", err)
			return
		}
		//处理订阅消息
		if response[0] == '*' {
			cnt, err := strconv.ParseInt(response[1:len(response)-2], 10, 32)
			if err != nil {
				fmt.Println("解析multibulk失败:", err)
				return
			}
			if int(cnt) == 3 {
				_, _ = reader.ReadString('\n')
				_, _ = reader.ReadString('\n')
				_, _ = reader.ReadString('\n')
				_, _ = reader.ReadString('\n')
				_, _ = reader.ReadString('\n')
				message, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("读取消息内容失败:", err)
					continue
				}
				fmt.Println("收到消息:", message[:len(message)-2])
			}
		}
	}
}
