package parser

// "*1\r\n$4\r\nPING\r\n"
// 第一个数字为 args 的数量， 接下来第一行代表下一行的字节数，第二行代表arg，因此两行表示一个arg

func Parse(lines [][]byte) [][]byte {
	lineCount := len(lines) // must be even
	args := make([][]byte, lineCount/2)
	for i := 0; i*2+1 < lineCount; i++ {
		args[i] = lines[2*i+1]
	}
	return args
}
