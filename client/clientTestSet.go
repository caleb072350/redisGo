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

	testCommand(conn, "*3\r\n$4\r\nSADD\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	// testCommand(conn, "*2\r\n$4\r\nEXISTS\r\n$3\r\nkey\r\n")
	// testCommand(conn, "*2\r\n$3\r\nDEL\r\n$3\r\nkey\r\n")
	// testCommand(conn, "*2\r\n$3\r\nEXISTS\r\n$3\r\nkey\r\n")
	// testCommand(conn, "*2\r\n$4\r\nTYPE\r\n$3\r\nkey\r\n")
	// testCommand(conn, "*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$8\r\nelement1\r\n$8\r\nelement2\r\n") // element1 -> element2
	// testCommand(conn, "*2\r\n$5\r\nTYPE\r\n$6\r\nmylist\r\n")
	// testCommand(conn, "*4\r\n$4\r\nHSET\r\n$7\r\nhashkey\r\n$5\r\nfield\r\n$5\r\nvalue\r\n")
	// testCommand(conn, "*2\r\n$5\r\nTYPE\r\n$7\r\nhashkey\r\n")
	// testCommand(conn, "*3\r\n$6\r\nRENAME\r\n$3\r\nkey\r\n$6\r\nnewkey\r\n")
	// testCommand(conn, "*3\r\n$6\r\nRENAMENX\r\n$3\r\nnewkey\r\n$6\r\nnewkey2\r\n")
	// testCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")

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
