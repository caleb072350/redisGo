# redis Go ~~~

## Running

You can get runnable program in the releases of this repository, which supports Linux and Darwin system.

```bash
./godis-darwin
```

You could use redis-cli or other redis client to connect godis server, which listens on 127.0.0.1:6379 on default mode.

The program will try to read config file path from environment variable `CONFIG`.

If environment variable is not set, then the program  try to read `redis.conf` in the working directory. 

If there is no such file, then the program will run with default config.

### cluster mode

Godis can work in cluster mode, please append following lines to redis.conf file

```ini
peers localhost:7379,localhost:7389 // other node in cluster
self  localhost:6399 // self address
```

We provide node1.conf and node2.conf for demonstration. 
use following command line to start a two-node-cluster:

```bash
CONFIG=node1.conf ./godis-darwin &
CONFIG=node2.conf ./godis-darwin &
``` 

The cluster is transparent to client. You can connect to any node in the cluster to access all data in the cluster:

```cmd
redis-cli -p 6399
```

## Commands

This repository implemented most of features of redis, including 5 kind of data structures, ttl, publish/subscribe and AOF persistence.

Supported Commands:

- Keys
    - del
    - expire
    - expireat
    - pexpire
    - pexpireat
    - ttl
    - pttl
    - persist
    - exists
    - type
    - rename
    - renamenx
- Server
    - flushdb
    - flushall
    - keys
    - bgrewriteaof
- String
    - set
    - setnx
    - setex
    - psetex
    - mset
    - mget
    - msetnx
    - get
    - getset
    - incr
    - incrby
    - incrbyfloat
    - decr
    - decrby
- List
    - lpush
    - lpushx
    - rpush
    - rpushx
    - lpop
    - rpop
    - rpoplpush
    - lrem
    - llen
    - lindex
    - lset
    - lrange
- Hash
    - hset
    - hsetnx
    - hget
    - hexists
    - hdel
    - hlen
    - hmget
    - hmset
    - hkeys
    - hvals
    - hgetall
    - hincrby
    - hincrbyfloat
- Set
    - sadd
    - sismember
    - srem
    - scard
    - smembers
    - sinter
    - sinterstore
    - sunion
    - sunionstore
    - sdiff
    - sdiffstore
    - srandmember
- SortedSet
    - zadd
    - zscore
    - zincrby
    - zrank
    - zcount
    - zrevrank
    - zcard
    - zrange
    - zrevrange
    - zrangebyscore
    - zrevrangebyscore
    - zrem
    - zremrangebyscore
    - zremrangebyrank
- Pub / Sub
    - publish
    - subscribe
    - unsubscribe

# Read My Code

If you want to read my code in this repository, here is a simple guidance.

- cmd: only the entry point
- config: config parser 
- interface: some interface definitions
- lib: some utils, such as logger, sync utils and wildcard

I suggest focusing on the following directories:

- server: the tcp server
- redis: the redis protocol parser
- datastruct: the implements of data structures
    - dict: a concurrent hash map
    - list: a linked list
    - lock: it is used to lock keys to ensure thread safety
    - set: a hash set based on map
    - sortedset: a sorted set implements based on skiplist
- db: the implements of the redis db
    - db.go: the basement of database
    - router.go: it find handler for commands 
    - keys.go: handlers for keys commands
    - string.go: handlers for string commands
    - list.go: handlers for list commands
    - hash.go: handlers for hash commands
    - set.go: handlers for set commands
    - sortedset.go: handlers for sorted set commands
    - pubsub.go: implements of publish / subscribe
    - aof.go: implements of AOF persistence and rewrite
