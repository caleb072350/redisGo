package db

import (
	"fmt"
	"os"
	"redisGo/config"
	Dict "redisGo/datastruct/dict"
	"redisGo/datastruct/lock"
	"redisGo/interface/dict"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/lib/timewheel"
	"redisGo/pubsub"
	"redisGo/redis/reply"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type DataEntity struct {
	Data interface{}
}

const (
	dataDictSize = 128
	ttlDictSize  = 128
	lockerSize   = 128
	aofQueueSize = 1 << 10
)

type CmdFunc func(db *DB, args [][]byte) redis.Reply

type DB struct {
	Data     dict.Dict
	TTLMap   dict.Dict
	Locker   *lock.LockMap
	interval time.Duration

	hub *pubsub.Hub

	stopWorld sync.WaitGroup // DB 的全局锁，在某些场景下单独对某个key加锁是不够的

	// main goroutine send command to aof goroutine through aofChan
	aofChan     chan *reply.MultiBulkReply
	aofFile     *os.File
	aofFilename string

	aofRewriteChan chan *reply.MultiBulkReply
	pausingAof     sync.RWMutex
}

/*
 * 将sync.RWMutex替换为sync.WaitGroup 的主要好处在于性能提升，尤其是在读多写少的场景下。
 * sync.RWMutex 允许多个读取操作并发执行，但写入操作会阻塞所有的读操作。虽然在一定程度上能够
 * 并发的读，提高了并发性读写锁本身也有一定的开销，每次获取和释放锁都需要进行系统调用，这会带来
 * 一定的开销 sync.WaitGroup则是一种更轻量级的同步机制，它主要用于等待一组goroutine完成执行。
 * 在Godis的场景中，主要实现了以下优化：
 * - 读取操作无锁：读取操作不再需要获取锁，可以直接访问Data，从而减少了锁的开销，提高了读取性能。
 * - 写入操作更高效：写入操作只需要等待所有进行的读操作完成后执行，而不需要阻塞后续的读取操作，这
 *   在读多写少的场景下可以显著提高吞吐量。
 * 具体实现方式：
 * - 在每个写操作前，调用wg.Add(1)增加计数器
 * - 在每个写操作结束后，调用wg.Done()减少计数器
 * - 在读操作开始前，调用wg.Wait()等待所有写操作完成
 */

var router = MakeRouter()

func MakeDB() *DB {
	db := &DB{
		Data:     Dict.MakeConcurrent(dataDictSize),
		TTLMap:   Dict.MakeConcurrent(ttlDictSize),
		Locker:   lock.Make(lockerSize),
		interval: 5 * time.Second,
		hub:      pubsub.MakeHub(),
	}

	if config.Properties.AppendOnly {
		db.aofFilename = config.Properties.AppendFilename
		db.loadAof(0)
		aofFile, err := os.OpenFile(db.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			logger.Warn(err)
		} else {
			db.aofFile = aofFile
			db.aofChan = make(chan *reply.MultiBulkReply, aofQueueSize)
		}
		go func() {
			db.handleAof()
		}()
	}
	db.TimerTask()
	return db
}

func (db *DB) Close() {
	if db.aofFile != nil {
		err := db.aofFile.Close()
		if err != nil {
			logger.Warn(err)
		}
	}
}

func (db *DB) Exec(c redis.Connection, args [][]byte) (result redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmd := strings.ToLower(string(args[0]))

	if cmd == "subscribe" {
		if len(args) < 2 {
			return &reply.ArgNumErrReply{Cmd: "subscribe"}
		}
		return pubsub.Subscribe(db.hub, c, args[1:])
	} else if cmd == "unsubscribe" {
		return pubsub.UnSubscribe(db.hub, c, args[1:])
	} else if cmd == "publish" {
		return pubsub.Publish(db.hub, args[1:])
	} else if cmd == "bgrewriteaof" {
		return BGRewriteAOF(db, args[1:])
	}
	CmdFunc, ok := router[cmd]
	if !ok {
		return reply.MakeErrReply("ERR unknown command `" + cmd + "`")
	}
	if len(args) > 1 {
		result = CmdFunc(db, args[1:])
	} else {
		result = CmdFunc(db, [][]byte{})
	}
	return
}

/* ---- Data Access ---- */
func (db *DB) Get(key string) (*DataEntity, bool) {
	db.stopWorld.Wait()

	raw, exists := db.Data.Get(key)
	if !exists {
		return nil, false
	}
	if db.IsExpired(key) {
		return nil, false
	}
	entity, _ := raw.(*DataEntity)
	return entity, true
}

func (db *DB) Put(key string, entity *DataEntity) int {
	db.stopWorld.Wait()
	return db.Data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *DataEntity) int {
	db.stopWorld.Wait()
	return db.Data.PutIfExists(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *DataEntity) int {
	db.stopWorld.Wait()
	return db.Data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) {
	db.stopWorld.Wait()
	db.Data.Remove(key)
	db.TTLMap.Remove(key)
}

func (db *DB) Removes(keys ...string) int {
	db.stopWorld.Wait()
	deleted := 0
	for _, key := range keys {
		_, exists := db.Data.Get(key)
		if exists {
			db.Data.Remove(key)
			db.TTLMap.Remove(key)
			deleted++
		}
	}
	return deleted
}

// 之前的stopWait.Wait()操作都是为了与Flush命令互斥访问数据
func (db *DB) Flush() {
	db.stopWorld.Add(1)
	defer db.stopWorld.Done()

	db.Data = Dict.MakeConcurrent(dataDictSize)
	db.TTLMap = Dict.MakeConcurrent(ttlDictSize)
	db.Locker = lock.Make(lockerSize)
}

/* ---- TTL Functions ---- */
func genExpireTask(key string) string {
	return "expire:" + key
}

func (db *DB) Expire(key string, expireTime time.Time) {
	db.stopWorld.Wait()
	db.TTLMap.Put(key, expireTime)
	taskKey := genExpireTask(key)
	timewheel.At(expireTime, taskKey, func() {
		logger.Info("expire: " + key)
		db.TTLMap.Remove(key)
		db.Data.Remove(key)
	})
}

func (db *DB) Persist(key string) {
	db.stopWorld.Wait()
	db.TTLMap.Remove(key)
	taskKey := genExpireTask(key)
	timewheel.Cancel(taskKey)
}

func (db *DB) IsExpired(key string) bool {
	rawExpireTime, ok := db.TTLMap.Get(key)
	if !ok {
		return false
	}
	expireTime, _ := rawExpireTime.(time.Time)
	expired := expireTime.Before(time.Now())
	return expired
}

func (db *DB) TimerTask() {

}

/* ---- Lock Function ---------------*/

func (db *DB) Lock(key string) {
	db.Locker.Lock(key)
}

func (db *DB) RLock(key string) {
	db.Locker.RLock(key)
}

func (db *DB) Unlock(key string) {
	db.Locker.Unlock(key)
}

func (db *DB) RUnlock(key string) {
	db.Locker.RUnlock(key)
}

func (db *DB) Locks(keys ...string) {
	db.Locker.Locks(keys...)
}

func (db *DB) RLocks(keys ...string) {
	db.Locker.RLocks(keys...)
}

func (db *DB) Unlocks(keys ...string) {
	db.Locker.Unlocks(keys...)
}

func (db *DB) RUnlocks(keys ...string) {
	db.Locker.RUnlocks(keys...)
}

func (db *DB) AfterClientClose(c redis.Connection) {
	pubsub.UnsubscribeAll(db.hub, c)
}
