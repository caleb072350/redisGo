// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	"strconv"
// )

// func main() {
// 	conn, err := net.Dial("tcp", "127.0.0.1:6379")
// 	if err != nil {
// 		fmt.Println("连接失败:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	testCommand(conn, "*4\r\n$4\r\nHSET\r\n$3\r\nkey\r\n$5\r\nfield\r\n$5\r\nvalue\r\n")
// 	testCommand(conn, "*4\r\n$6\r\nHSETNX\r\n$3\r\nkey\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n")
// 	testCommand(conn, "*3\r\n$4\r\nHGET\r\n$3\r\nkey\r\n$5\r\nfield\r\n")
// 	testCommand(conn, "*3\r\n$4\r\nHGET\r\n$3\r\nkey\r\n$6\r\nfield2\r\n")
// 	testCommand(conn, "*3\r\n$6\r\nHEXISTS\r\n$3\r\nkey\r\n$5\r\nfield\r\n")
// 	testCommand(conn, "*3\r\n$4\r\nHDEL\r\n$3\r\nkey\r\n$5\r\nfield\r\n")
// 	testCommand(conn, "*3\r\n$6\r\nHEXISTS\r\n$3\r\nkey\r\n$5\r\nfield\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nHLEN\r\n$3\r\nkey\r\n")
// 	testCommand(conn, "*3\r\n$4\r\nHDEL\r\n$3\r\nkey\r\n$5\r\nfield2\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nHLEN\r\n$3\r\nkey\r\n")

// 	testCommand(conn, "*6\r\n$4\r\nHMSET\r\n$3\r\nkey\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n")

// 	testCommand(conn, "*3\r\n$4\r\nHGET\r\n$3\r\nkey\r\n$6\r\nfield2\r\n")
// 	testCommand(conn, "*4\r\n$5\r\nHMGET\r\n$3\r\nkey\r\n$6\r\nfield1\r\n$6\r\nfield2\r\n")

// 	testCommand(conn, "*2\r\n$5\r\nHKEYS\r\n$3\r\nkey\r\n")
// 	testCommand(conn, "*2\r\n$5\r\nHVALS\r\n$3\r\nkey\r\n")
// 	testCommand(conn, "*2\r\n$7\r\nHGETALL\r\n$3\r\nkey\r\n")

// 	testCommand(conn, "*4\r\n$4\r\nHSET\r\n$3\r\nkey\r\n$5\r\nfield\r\n$1\r\n5\r\n")
// 	testCommand(conn, "*4\r\n$7\r\nHINCRBY\r\n$3\r\nkey\r\n$5\r\nfield\r\n$1\r\n5\r\n")
// 	testCommand(conn, "*4\r\n$12\r\nHINCRBYFLOAT\r\n$3\r\nkey\r\n$5\r\nfield\r\n$3\r\n1.5\r\n")
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
