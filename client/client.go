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

// 	// 测试 SET 命令 String
// 	testCommand(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n") // 设置 key-value
// 	// 测试 GET 命令
// 	testCommand(conn, "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n") // 获取 key 的值
// 	// 测试list的rpush命令,如果list不存在，则新建list，并加入元素
// 	testCommand(conn, "*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$8\r\nelement1\r\n$8\r\nelement2\r\n") // element1 -> element2
// 	// 测试list的lindex命令
// 	testCommand(conn, "*3\r\n$6\r\nLINDEX\r\n$6\r\nmylist\r\n$1\r\n1\r\n") // element2
// 	// 测试list的LLen命令
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n") // 2
// 	// 测试list的LPop命令
// 	testCommand(conn, "*2\r\n$4\r\nlpop\r\n$6\r\nmylist\r\n") // element1
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n") // 1
// 	// 测试list的LPush命令
// 	testCommand(conn, "*4\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$8\r\nelement3\r\n$8\r\nelement4\r\n") // element3 -> element4 -> element2
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n")                                      // 3
// 	testCommand(conn, "*3\r\n$6\r\nLINDEX\r\n$6\r\nmylist\r\n$1\r\n1\r\n")                         // element4
// 	// 测试list的LRange命令
// 	testCommand(conn, "*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n") // element4 -> element3 -> element2
// 	// 测试list的LRem命令
// 	testCommand(conn, "*4\r\n$5\r\nLREM\r\n$6\r\nmylist\r\n$1\r\n1\r\n$8\r\nelement3\r\n") // element4 -> element2
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n")                              //2
// 	testCommand(conn, "*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n")      // element4 -> element2

// 	// 测试list的LSet命令
// 	testCommand(conn, "*4\r\n$4\r\nlset\r\n$6\r\nmylist\r\n$1\r\n1\r\n$8\r\nelement5\r\n") // element4 -> element5
// 	testCommand(conn, "*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n3\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n")

// 	// 测试list的RPop命令
// 	testCommand(conn, "*2\r\n$4\r\nrpop\r\n$6\r\nmylist\r\n")              // element5
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n")              // 1
// 	testCommand(conn, "*3\r\n$6\r\nLINDEX\r\n$6\r\nmylist\r\n$1\r\n0\r\n") // element4

// 	testCommand(conn, "*4\r\n$5\r\nRPUSH\r\n$7\r\nmylist2\r\n$8\r\nelement1\r\n$8\r\nelement2\r\n")
// 	// 测试list的RPopLPush命令
// 	testCommand(conn, "*3\r\n$9\r\nRPOPLPUSH\r\n$6\r\nmylist\r\n$7\r\nmylist2\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$7\r\nmylist2\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nllen\r\n$6\r\nmylist\r\n")

// 	//测试string的Del命令
// 	testCommand(conn, "*2\r\n$3\r\ndel\r\n$3\r\nkey\r\n")
// 	// testCommand(conn, "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n")
// 	// 测试string的MSet命令
// 	testCommand(conn, "*5\r\n$4\r\nmset\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n$4\r\nkey2\r\n$6\r\nvalue2\r\n")
// 	testCommand(conn, "*2\r\n$3\r\nget\r\n$4\r\nkey1\r\n")
// 	testCommand(conn, "*2\r\n$3\r\nget\r\n$4\r\nkey2\r\n")
// 	testCommand(conn, "*3\r\n$4\r\nmget\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n")

// 	// 测试string的GetSet命令
// 	testCommand(conn, "*3\r\n$3\r\nset\r\n$4\r\nkey1\r\n$1\r\n3\r\n")
// 	testCommand(conn, "*3\r\n$6\r\ngetset\r\n$4\r\nkey1\r\n$1\r\n4\r\n")
// 	testCommand(conn, "*2\r\n$4\r\nincr\r\n$4\r\nkey1\r\n") // 5

// 	// 测试string的IncrBy命令
// 	testCommand(conn, "*3\r\n$4\r\nincrby\r\n$4\r\nkey1\r\n$1\r\n3\r\n")

// 	// 测试string的IncrByFloat命令
// 	// testCommand(conn, "*3\r\n$12\r\nincrbyfloat\r\n$4\r\nkey1\r\n$3\r\n3.2\r\n")

// 	// 测试string的Decr命令
// 	testCommand(conn, "*2\r\n$4\r\ndecr\r\n$4\r\nkey1\r\n")
// 	testCommand(conn, "*3\r\n$6\r\ndecrby\r\n$4\r\nkey1\r\n$1\r\n3\r\n")
// 	testCommand(conn, "*3\r\n$11\r\ndecrbyfloat\r\n$4\r\nkey1\r\n$3\r\n3.2\r\n")

// 	time.Sleep(time.Second)
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