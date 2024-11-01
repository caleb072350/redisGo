package cluster

import (
	"context"
	"errors"
	"fmt"
	"redisGo/cluster/idgenerator"
	"redisGo/config"
	"redisGo/datastruct/dict"
	"redisGo/db"
	"redisGo/interface/redis"
	"redisGo/lib/consistenthash"
	"redisGo/lib/logger"
	"redisGo/redis/client"
	"redisGo/redis/reply"
	"runtime/debug"
	"strings"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type Cluster struct {
	self           string
	db             *db.DB
	peerPicker     *consistenthash.Map
	peerConnection map[string]*pool.ObjectPool

	transactions *dict.SimpleDict // id -> Transaction
	idGenerator  *idgenerator.IdGenerator
}

const (
	replicas = 4
)

func MakeCluster() *Cluster {
	cluster := &Cluster{
		self:           config.Properties.Self,
		db:             db.MakeDB(),
		peerPicker:     consistenthash.New(replicas, nil),
		peerConnection: make(map[string]*pool.ObjectPool),

		transactions: dict.MakeSimple(),
		idGenerator:  idgenerator.MakeGenerator("godis", config.Properties.Self),
	}
	if config.Properties.Peers != nil && len(config.Properties.Peers) > 0 && config.Properties.Self != "" {
		contains := make(map[string]bool)
		peers := make([]string, 0, len(config.Properties.Peers)+1)
		for _, peer := range config.Properties.Peers {
			if _, ok := contains[peer]; ok {
				continue
			}
			contains[peer] = true
			peers = append(peers, peer)
		}
		peers = append(peers, config.Properties.Self)
		cluster.peerPicker.Add(peers...)
		ctx := context.Background()
		for _, peer := range peers {
			cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &ConnectionFactory{Peer: peer})
		}
	}
	return cluster
}

type CmdFunc func(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply

func (cluster *Cluster) Close() {
	cluster.db.Close()
}

var router = MakeRouter()

func (cluster *Cluster) Exec(c redis.Connection, args [][]byte) (result redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmd := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmd]
	if !ok {
		return reply.MakeErrReply("ERR unknown command `" + cmd + "`")
	}
	result = cmdFunc(cluster, c, args)
	return
}

// replay command to peer
// cannot call Prepare, Commit, Rollback of self node
func (cluster *Cluster) Relay(peer string, c redis.Connection, args [][]byte) redis.Reply {
	if peer == cluster.self {
		return cluster.db.Exec(c, args)
	} else {
		peerClient, err := cluster.getPeerClient(peer)
		// lazy init
		if err != nil {
			return reply.MakeErrReply(err.Error())
		}
		defer func() {
			_ = cluster.returnPeerClient(peer, peerClient)
		}()
		return peerClient.Send(args)
	}
}

func (cluster *Cluster) AfterClientClose(c redis.Connection) {

}

func (cluster *Cluster) getPeerClient(peer string) (*client.Client, error) {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection factory not found")
	}
	raw, err := connectionFactory.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(*client.Client)
	if !ok {
		return nil, errors.New("connection factory make wrong type")
	}
	return conn, nil
}

func (cluster *Cluster) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}
func Ping(cluster *Cluster, r redis.Connection, args [][]byte) redis.Reply {
	if len(args) == 1 {
		return &reply.PongReply{}
	} else if len(args) == 2 {
		return reply.MakeStatusReply("\"" + string(args[1]) + "\"")
	} else {
		return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	}
}

/*----- utils -------*/

func makeArgs(cmd string, args ...string) [][]byte {
	result := make([][]byte, len(args)+1)
	result[0] = []byte(cmd)
	for i, arg := range args {
		result[i+1] = []byte(arg)
	}
	return result
}

// return peer -> keys
func (cluster *Cluster) groupBy(keys []string) map[string][]string {
	result := make(map[string][]string)
	for _, key := range keys {
		peer := cluster.peerPicker.Get(key)
		group, ok := result[peer]
		if !ok {
			group = make([]string, 0)
		}
		group = append(group, key)
		result[peer] = group
	}
	return result
}
