// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	"strconv"
// 	"time"
// )

// func main() {
// 	conn, err := net.Dial("tcp", "127.0.0.1:6379")
// 	if err != nil {
// 		fmt.Println("连接失败:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// 测试string的Expire命令, 过期时间为 10s
// 	// testCommand(conn, "*5\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$2\r\nEX\r\n$2\r\n10\r\n")
// 	// time.Sleep(11 * time.Second)
// 	// 判断是否过期
// 	// testCommand(conn, "*2\r\n$3\r\nISEXPIRED\r\n$3\r\nkey\r\n")
// 	testCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
// 	// testCommand(conn, "*3\r\n$3\r\nEXPIRE\r\n$3\r\nkey\r\n$1\r\n3\r\n")
// 	timestamp := time.Now().Add(5 * time.Second).Unix()
// 	testCommand(conn, "*3\r\n$8\r\nEXPIREAT\r\n$3\r\nkey\r\n$10\r\n"+strconv.FormatInt(timestamp, 10)+"\r\n")
// 	// 设置key 5s之后过期，sleep 1s，然后查询key的ttl
// 	time.Sleep(1 * time.Second)
// 	testCommand(conn, "*2\r\n$7\r\nPERSIST\r\n$3\r\nkey\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nPTTL\r\n$3\r\nkey\r\n")
// }

// func testCommand(conn net.Conn, command string) {
// 	fmt.Printf("发送命令:\n%s", command)

// 	// 发送命令
// 	fmt.Fprint(conn, command)
// 	reader := bufio.NewReader(conn)
// 	response, err := reader.ReadString('\n')
// 	if err != nil {
// 		fmt.Println("读取响应失败:", err)
// 		return
// 	}
// 	fmt.Println("响应:", response)
// 	if response[0] == '$' { // bulk response
// 		response, err = reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("读取响应失败:", err)
// 			return
// 		}
// 		fmt.Println("响应:", response)
// 	} else if response[0] == '*' {
// 		cnt, err := strconv.ParseInt(response[1:len(response)-2], 10, 32)
// 		if err != nil {
// 			fmt.Println("解析multibulk协议失败:", err)
// 			return
// 		}
// 		for i := 0; i < int(cnt*2); i++ {
// 			response, err = reader.ReadString('\n')
// 			if err != nil {
// 				fmt.Println("读取响应失败:", err)
// 				return
// 			}
// 			fmt.Println("响应:", response)
// 		}
// 	}

// }
