package db

import (
	"fmt"
	"redisGo/datastruct/dict"
	"redisGo/datastruct/lock"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"runtime/debug"
	"strings"
	"time"
)

type DataEntity struct {
	Data interface{}
}

type CmdFunc func(db *DB, args [][]byte) redis.Reply

type DB struct {
	Data     *dict.ConcurrentDict
	TTLMap   *dict.SimpleDict
	Locker   *lock.LockMap
	interval time.Duration
}

var router = MakeRouter()

func (db *DB) Exec(args [][]byte) (result redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmd := strings.ToLower(string(args[0]))
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

func (db *DB) Get(key string) (*DataEntity, bool) {
	raw, exists := db.Data.Get(key)
	if !exists {
		return nil, false
	}
	entity, _ := raw.(*DataEntity)
	return entity, true
}

func (db *DB) Remove(key string) {
	db.Data.Remove(key)
	db.TTLMap.Remove(key)
}

func (db *DB) Expire(key string, expireTime time.Time) {
	db.TTLMap.Put(key, expireTime)
}

func (db *DB) Persist(key string) {
	db.TTLMap.Remove(key)
}

func (db *DB) IsExpired(key string) bool {
	rawExpireTime, ok := db.TTLMap.Get(key)
	if !ok {
		return true
	}
	expireTime, _ := rawExpireTime.(time.Time)
	expired := expireTime.Before(time.Now())
	return expired
}

func (db *DB) CleanExpired() {
	now := time.Now()
	db.TTLMap.ForEach(func(key string, value interface{}) bool {
		expireTime, _ := value.(time.Time)
		if expireTime.Before(now) {
			db.Remove(key)
			logger.Info(fmt.Sprintf("clean expired key: %s", key))
		}
		return true
	})
}

func (db *DB) TimerTask() {
	ticker := time.NewTicker(db.interval)
	go func() {
		for range ticker.C {
			db.CleanExpired()
			logger.Info("TimerTask is running...")
		}
	}()
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

func MakeDB() *DB {
	db := &DB{
		Data:     dict.MakeConcurrent(128),
		TTLMap:   dict.MakeSimple(),
		Locker:   lock.Make(1024),
		interval: 5 * time.Second,
	}
	db.TimerTask()
	return db
}
