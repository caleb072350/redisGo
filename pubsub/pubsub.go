package pubsub

import (
	"redisGo/datastruct/list"
	"redisGo/interface/redis"
	"redisGo/redis/reply"
	"strconv"
)

var (
	_subscribe         = "subscribe"
	_unsubscribe       = "unsubscribe"
	messageBytes       = []byte("message")
	unSubscribeNothing = []byte("*3\r\n$11\r\nunsubscribe\r\n$-1\n:0\r\n")
)

func MakeMsg(t string, channel string, code int64) []byte {
	return []byte("*3\r\n$" + strconv.FormatInt(int64(len(t)), 10) + reply.CRLF + t + reply.CRLF +
		"$" + strconv.FormatInt(int64(len(channel)), 10) + reply.CRLF + channel + reply.CRLF +
		":" + strconv.FormatInt(code, 10) + reply.CRLF)
}

/*
 * invoker should lock channel
 */
func subscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.SubsChannel(channel)

	// add into db.subs
	raw, ok := hub.subs.Get(channel)
	var subscribers *list.LinkedList
	if ok {
		subscribers, _ = raw.(*list.LinkedList)
	} else {
		subscribers = list.Make()
		hub.subs.Put(channel, subscribers)
	}
	if subscribers.Contains(client) {
		return false
	}
	subscribers.Add(client)
	return true
}

func unsubscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.UnSubsChannel(channel)

	raw, ok := hub.subs.Get(channel)
	if ok {
		subscribers, _ := raw.(*list.LinkedList)
		subscribers.RemoveAllByVal(client)
		if subscribers.Len() == 0 {
			hub.subs.Remove(channel)
		}
		return true
	}
	return false
}

func Subscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {
	channels := make([]string, len(args))
	for i, b := range args {
		channels[i] = string(b)
	}
	hub.subsLocker.Locks(channels...)
	defer hub.subsLocker.Unlocks(channels...)

	for _, channel := range channels {
		if subscribe0(hub, channel, c) {
			_ = c.Write(MakeMsg(_subscribe, channel, int64(c.SubsCount())))
		}
	}
	return &reply.NoReply{}
}

func UnsubscribeAll(hub *Hub, c redis.Connection) {
	channels := c.GetChannels()
	hub.subsLocker.Locks(channels...)
	defer hub.subsLocker.Unlocks(channels...)

	for _, channel := range channels {
		unsubscribe0(hub, channel, c)
	}
}

func UnSubscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {
	var channels []string
	if len(args) > 0 {
		channels = make([]string, len(args))
		for i, b := range args {
			channels[i] = string(b)
		}
	} else {
		channels = c.GetChannels()
	}
	hub.subsLocker.Locks(channels...)
	defer hub.subsLocker.Unlocks(channels...)

	if len(channels) == 0 {
		_ = c.Write(unSubscribeNothing)
		return &reply.NoReply{}
	}

	for _, channel := range channels {
		if unsubscribe0(hub, channel, c) {
			_ = c.Write(MakeMsg(_unsubscribe, channel, int64(c.SubsCount())))
		}
	}
	return &reply.NoReply{}
}

func Publish(hub *Hub, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return &reply.ArgNumErrReply{Cmd: "publish"}
	}
	channel := string(args[0])
	message := args[1]

	hub.subsLocker.Lock(channel)
	defer hub.subsLocker.Unlock(channel)

	raw, ok := hub.subs.Get(channel)
	if !ok {
		return reply.MakeIntReply(0)
	}
	subscribers, _ := raw.(*list.LinkedList)
	subscribers.ForEach(func(_ int, c interface{}) bool {
		client, _ := c.(redis.Connection)
		replyArgs := make([][]byte, 3)
		replyArgs[0] = messageBytes
		replyArgs[1] = []byte(channel)
		replyArgs[2] = message
		_ = client.Write(reply.MakeMultiBulkReply(replyArgs).ToBytes())
		return true
	})
	return reply.MakeIntReply(int64(subscribers.Len()))
}
