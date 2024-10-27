package db

import (
	"bufio"
	"io"
	"os"
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"strconv"
	"strings"
)

func makeAofCmd(cmd string, args [][]byte) *reply.MultiBulkReply {
	params := make([][]byte, len(args)+1)
	copy(params[1:], args)
	params[0] = []byte(cmd)
	return reply.MakeMultiBulkReply(params)
}

func (db *DB) addAof(args *reply.MultiBulkReply) {
	db.aofChan <- args
}

func (db *DB) handleAof() {
	for cmd := range db.aofChan {
		_, err := db.aofFile.Write(cmd.ToBytes())
		if err != nil {
			logger.Warn(err)
		}
	}
}

func (db *DB) loadAof() {
	file, err := os.Open(db.aofFilename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var fixedLen int64 = 0
	var expectedArgsCount uint32
	var receivedCount uint32
	var args [][]byte
	processing := false
	var msg []byte
	for {
		if fixedLen == 0 {
			msg, err := reader.ReadBytes('\n')
			if err == io.EOF {
				return
			}
			if len(msg) < 2 || msg[len(msg)-2] != '\r' {
				logger.Warn("invalid format: line should end with '\r\n'")
				return
			}
		} else {
			msg := make([]byte, fixedLen+2)
			_, err = io.ReadFull(reader, msg)
			if err == io.EOF {
				return
			}
			if len(msg) == 0 || msg[len(msg)-1] != '\n' || msg[len(msg)-2] != '\r' {
				logger.Warn("invalid multibulk length")
				return
			}
			fixedLen = 0
		}
		if err != nil {
			logger.Warn(err)
			return
		}
		if !processing {
			if msg[0] == '*' {
				expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					logger.Warn(err)
					return
				}
				expectedArgsCount = uint32(expectedLine)
				receivedCount = 0
				processing = true
				args = make([][]byte, expectedLine)
			} else {
				logger.Warn("msg should start with '*'")
				return
			}
		} else {
			line := msg[0 : len(msg)-2]
			if line[0] == '$' {
				fixedLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
				if err != nil {
					logger.Warn(err)
					return
				}
				if fixedLen <= 0 {
					logger.Warn("invalid multibulk length")
					return
				}
			} else {
				args[receivedCount] = line
				receivedCount++
			}

			// if sending finished
			if receivedCount == expectedArgsCount {
				processing = false
				cmd := strings.ToLower(string(args[0]))
				cmdFunc, ok := router[cmd]
				if ok {
					cmdFunc(db, args[1:])
				}

				// finish
				expectedArgsCount = 0
				receivedCount = 0
				args = nil
			}
		}
	}
}
