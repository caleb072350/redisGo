package db

import (
	"bufio"
	"io"
	"os"
	"redisGo/config"
	Dict "redisGo/datastruct/dict"
	"redisGo/datastruct/list"
	"redisGo/datastruct/lock"
	"redisGo/datastruct/set"
	"redisGo/interface/dict"
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"strconv"
	"strings"
	"time"
)

var pExpireAtCmd = []byte("PEXPIREAT")

func makeExpireCmd(key string, expireAt time.Time) *reply.MultiBulkReply {
	args := make([][]byte, 3)
	args[0] = pExpireAtCmd
	args[1] = []byte(key)
	args[2] = []byte(strconv.FormatInt(int64(expireAt.UnixNano()/1e6), 10))
	return reply.MakeMultiBulkReply(args)
}

func makeAofCmd(cmd string, args [][]byte) *reply.MultiBulkReply {
	params := make([][]byte, len(args)+1)
	copy(params[1:], args)
	params[0] = []byte(cmd)
	return reply.MakeMultiBulkReply(params)
}

func (db *DB) AddAof(args *reply.MultiBulkReply) {
	if config.Properties.AppendOnly && db.aofChan != nil {
		db.aofChan <- args
	}
}

func (db *DB) handleAof() {
	for cmd := range db.aofChan {
		db.pausingAof.RLock()
		if db.aofRewriteChan != nil {
			db.aofRewriteChan <- cmd
		}
		_, err := db.aofFile.Write(cmd.ToBytes())
		if err != nil {
			logger.Warn(err)
		}
		db.pausingAof.RUnlock()
	}
}

func trim(msg []byte) string {
	trimed := ""
	for i := len(msg) - 1; i >= 0; i-- {
		if msg[i] == '\n' || msg[i] == '\r' {
			continue
		}
		return string(msg[:i+1])
	}
	return trimed
}

func (db *DB) loadAof(maxBytes int) {
	// delete aofChan to prevent write again
	aofChan := db.aofChan
	db.aofChan = nil
	defer func(aofChan chan *reply.MultiBulkReply) {
		db.aofChan = aofChan
	}(aofChan)

	// load aof
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
	readBytes := 0
	for {
		if maxBytes != 0 && readBytes > maxBytes {
			break
		}
		if fixedLen == 0 {
			msg, err := reader.ReadBytes('\n')
			if err == io.EOF {
				return
			}
			if len(msg) == 0 {
				logger.Warn("invalid format: line should end with \\r\\n")
				return
			}
			readBytes += len(msg)
		} else {
			msg := make([]byte, fixedLen+2)
			n, err := io.ReadFull(reader, msg)
			if err == io.EOF {
				return
			}
			if len(msg) == 0 {
				logger.Warn("invalid multibulk length")
				return
			}
			fixedLen = 0
			readBytes += n
		}
		if err != nil {
			logger.Warn(err)
			return
		}
		if !processing && len(msg) >= 2 {
			if msg[0] == '*' {
				expectedLine, err := strconv.ParseUint(trim(msg[1:]), 10, 32)
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
		} else if len(msg) >= 2 {
			line := msg[0 : len(msg)-2]
			if line[0] == '$' {
				fixedLen, err = strconv.ParseInt(trim(line[1:]), 10, 64)
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

/* aof rewrite 主要是aof文件很大之后影响读写性能，需要重写aof文件，重写aof文件会简化中间的操作过程，仅保证最终的数据一致 */
func (db *DB) aofRewrite() {
	file, fileSize, err := db.startRewrite()
	if err != nil {
		logger.Warn(err)
		return
	}

	// load aof file
	tmpDB := &DB{
		Data:        Dict.MakeConcurrent(dataDictSize),
		TTLMap:      Dict.MakeConcurrent(ttlDictSize),
		Locker:      lock.Make(lockerSize),
		interval:    5 * time.Second,
		aofFilename: db.aofFilename,
	}
	tmpDB.loadAof(int(fileSize))

	// rewrite aof file
	tmpDB.Data.ForEach(func(key string, raw interface{}) bool {
		var cmd *reply.MultiBulkReply
		entity, _ := raw.(*DataEntity)
		switch val := entity.Data.(type) {
		case []byte:
			cmd = persistString(key, val)
		case *list.LinkedList:
			cmd = persistList(key, val)
		case *set.Set:
			cmd = persistSet(key, val)
		case dict.Dict:
			cmd = persistHash(key, val)
		}
		if cmd != nil {
			_, _ = file.Write(cmd.ToBytes())
		}
		return true
	})

	tmpDB.TTLMap.ForEach(func(key string, raw interface{}) bool {
		expireTime, _ := raw.(time.Time)
		cmd := makeExpireCmd(key, expireTime)
		if cmd != nil {
			_, _ = file.Write(cmd.ToBytes())
		}
		return true
	})
	db.finishRewrite(file)
}

var setCmd = []byte("SET")

func persistString(key string, bytes []byte) *reply.MultiBulkReply {
	args := make([][]byte, 3)
	args[0] = setCmd
	args[1] = []byte(key)
	args[2] = bytes
	return reply.MakeMultiBulkReply(args)
}

var rPushAllCmd = []byte("RPUSHALL")

func persistList(key string, list *list.LinkedList) *reply.MultiBulkReply {
	args := make([][]byte, 2+list.Len())
	args[0] = rPushAllCmd
	args[1] = []byte(key)
	list.ForEach(func(i int, val interface{}) bool {
		bytes, _ := val.([]byte)
		args[2+i] = bytes
		return true
	})
	return reply.MakeMultiBulkReply(args)
}

var sAddCmd = []byte("SADD")

func persistSet(key string, set *set.Set) *reply.MultiBulkReply {
	args := make([][]byte, 2+set.Len())
	args[0] = sAddCmd
	args[1] = []byte(key)
	i := 0
	set.ForEach(func(member string) bool {
		args[2+i] = []byte(member)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(args)
}

var hMSetCmd = []byte("HMSET")

func persistHash(key string, hash dict.Dict) *reply.MultiBulkReply {
	args := make([][]byte, 2+hash.Len()*2)
	args[0] = hMSetCmd
	args[1] = []byte(key)
	i := 0
	hash.ForEach(func(member string, val interface{}) bool {
		args[2+i*2] = []byte(member)
		args[3+i*2] = val.([]byte)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(args)
}
func (db *DB) startRewrite() (*os.File, int64, error) {
	db.pausingAof.Lock() // pausing aof
	defer db.pausingAof.Unlock()

	// create rewrite channel
	db.aofRewriteChan = make(chan *reply.MultiBulkReply, aofQueueSize)

	// get current aof file size
	fileInfo, _ := os.Stat(db.aofFilename)
	filesize := fileInfo.Size()

	// create tmp file
	file, err := os.CreateTemp("", "aof")
	if err != nil {
		logger.Warn("tmp file create failed")
		return nil, 0, err
	}
	return file, filesize, nil
}

func (db *DB) finishRewrite(tmpFile *os.File) {
	db.pausingAof.Lock() // pausing aof
	defer db.pausingAof.Unlock()

	// 将执行rewriteAof过程中接收到的命令写入tmpFile
loop:
	for {
		select {
		case cmd := <-db.aofRewriteChan:
			_, err := tmpFile.Write(cmd.ToBytes())
			if err != nil {
				logger.Warn(err)
			}
		default:
			break loop
		}
	}

	close(db.aofRewriteChan)
	db.aofRewriteChan = nil

	// rename tmp file
	_ = db.aofFile.Close()
	_ = os.Rename(tmpFile.Name(), db.aofFilename)

	aofFile, err := os.OpenFile(db.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	db.aofFile = aofFile
}
