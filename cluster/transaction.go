package cluster

import (
	"fmt"
	"redisGo/db"
	"redisGo/interface/redis"
	"redisGo/lib/logger"
	"redisGo/lib/marshal/gob"
	"redisGo/lib/timewheel"
	"redisGo/redis/reply"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Transaction struct {
	id      string
	args    [][]byte
	cluster *Cluster
	conn    redis.Connection

	keys    []string
	undoLog map[string][]byte

	status int8
	mu     *sync.Mutex
}

const (
	maxLockTime       = 3 * time.Second
	waitBeforeCleanTx = 2 * time.Second
	CreatedStatus     = iota
	PreparedStatus
	CommittedStatus
	RollbackedStatus
)

func genTaskKey(txId string) string {
	return "tx:" + txId
}

func NewTransaction(cluster *Cluster, c redis.Connection, id string, args [][]byte, keys []string) *Transaction {
	return &Transaction{
		id:      id,
		args:    args,
		cluster: cluster,
		conn:    c,
		keys:    keys,
		status:  CreatedStatus,
		mu:      new(sync.Mutex),
	}
}

// t should contains keys field and Id field
func (tx *Transaction) prepare() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	// lock keys
	tx.cluster.db.Locks(tx.keys...)

	// build undoLog
	tx.undoLog = make(map[string][]byte)
	for _, key := range tx.keys {
		entity, ok := tx.cluster.db.Get(key)
		if ok {
			blob, err := gob.Marshal(entity)
			if err != nil {
				return err
			}
			tx.undoLog[key] = blob
		} else {
			tx.undoLog[key] = []byte{}
		}
	}
	tx.status = PreparedStatus
	taskKey := genTaskKey(tx.id)
	timewheel.Delay(maxLockTime, taskKey, func() {
		if tx.status == PreparedStatus { // rollback transaction uncommitted until expire
			logger.Info("abort transaction: " + tx.id)
			_ = tx.rollback()
		}
	})
	return nil
}

func (tx *Transaction) rollback() error {
	curStatus := tx.status
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.status != curStatus {
		return fmt.Errorf("tx %s status changed", tx.id)
	}
	for key, blob := range tx.undoLog {
		if len(blob) > 0 {
			entity := &db.DataEntity{}
			err := gob.UnMarshal(blob, entity)
			if err != nil {
				return err
			}
			tx.cluster.db.Put(key, entity)
		} else {
			tx.cluster.db.Remove(key)
		}
	}
	// a committed transaction has released locks, do not release again
	if tx.status != CommittedStatus {
		tx.cluster.db.Unlocks(tx.keys...)
	}
	tx.status = RollbackedStatus
	return nil
}

// rollback local transaction
func Rollback(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rollback' command")
	}
	txId := string(args[1])
	raw, ok := cluster.transactions.Get(txId)
	if !ok {
		return reply.MakeIntReply(0)
	}
	tx, _ := raw.(*Transaction)
	err := tx.rollback()
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	// clean transaction
	timewheel.Delay(waitBeforeCleanTx, "", func() {
		cluster.transactions.Remove(tx.id)
	})
	return reply.MakeIntReply(1)
}

func Commit(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'commit' command")
	}
	txId := string(args[1])
	raw, ok := cluster.transactions.Get(txId)
	if !ok {
		return reply.MakeIntReply(0)
	}
	tx, _ := raw.(*Transaction)

	tx.mu.Lock()
	defer tx.mu.Unlock()

	//finish transaction
	defer func() {
		cluster.db.Unlocks(tx.keys...)
		tx.status = CommittedStatus
	}()

	cmd := strings.ToLower(string(tx.args[0]))
	var result redis.Reply
	if cmd == "del" {
		result = CommitDel(cluster, c, tx)
	}
	if reply.IsErrorReply(result) {
		//failed
		err2 := tx.rollback()
		return reply.MakeErrReply(fmt.Sprintf("err occurs when rollback: %v. origin err: %v", err2, result))
	} else {
		// after committed
		cluster.db.Unlocks(tx.keys...)
		tx.status = CommittedStatus
		// clean transaction
		// do not clean immediately, in case rollback
		timewheel.Delay(waitBeforeCleanTx, "", func() {
			cluster.transactions.Remove(tx.id)
		})
	}
	return result
}

// request all node rollback transaction as leader
func RequestRollback(cluster *Cluster, c redis.Connection, txId int64, peers map[string][]string) {
	txIdStr := strconv.FormatInt(txId, 10)
	for peer := range peers {
		if peer == cluster.self {
			Rollback(cluster, c, makeArgs("rollback", txIdStr))
		} else {
			cluster.Relay(peer, c, makeArgs("rollback", txIdStr))
		}
	}
}

// request all node commit transaction as leader
func RequestCommit(cluster *Cluster, c redis.Connection, txId int64, peers map[string][]string) ([]redis.Reply, reply.ErrorReply) {
	var errReply reply.ErrorReply
	txIdStr := strconv.FormatInt(txId, 10)
	respList := make([]redis.Reply, 0, len(peers))
	for peer := range peers {
		var resp redis.Reply
		if peer == cluster.self {
			resp = Commit(cluster, c, makeArgs("commit", txIdStr))
		} else {
			resp = cluster.Relay(peer, c, makeArgs("commit", txIdStr))
		}
		if reply.IsErrorReply(resp) {
			errReply = resp.(reply.ErrorReply)
			break
		}
		respList = append(respList, resp)
	}
	if errReply != nil {
		RequestRollback(cluster, c, txId, peers)
		return nil, errReply
	}
	return respList, nil
}
