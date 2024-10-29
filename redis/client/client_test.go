package client

import (
	"redisGo/lib/logger"
	"redisGo/redis/reply"
	"testing"
)

func TestClient(t *testing.T) {
	logger.Setup(&logger.Settings{
		Path:       "./logs",
		Name:       "myGodis",
		Ext:        ".log",
		TimeFormat: "2006-01-02",
	})
	client, err := MakeClient("127.0.0.1:6379")
	if err != nil {
		t.Fatal(err)
	}
	client.Start()

	result := client.Send([][]byte{
		[]byte("PING"),
	})
	if statusRet, ok := result.(*reply.StatusReply); ok {
		if statusRet.Status != "PONG" {
			t.Error("`ping` failed, result: " + statusRet.Status)
		}
	}

	result = client.Send([][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("value"),
	})
	if statusRet, ok := result.(*reply.StatusReply); ok {
		if statusRet.Status != "OK" {
			t.Error("`set` failed, result: " + statusRet.Status)
		}
	}

	result = client.Send([][]byte{
		[]byte("GET"),
		[]byte("key"),
	})
	if bulkRet, ok := result.(*reply.BulkReply); ok {
		if string(bulkRet.Arg) != "value" {
			t.Error("`get` failed, result: " + string(bulkRet.Arg))
		}
	}

	result = client.Send([][]byte{
		[]byte("DEL"),
		[]byte("key"),
	})
	if intRet, ok := result.(*reply.IntReply); ok {
		if intRet.Code != 1 {
			t.Errorf("`del` failed, result: %d", intRet.Code)
		}
	}

	result = client.Send([][]byte{
		[]byte("GET"),
		[]byte("key"),
	})
	if _, ok := result.(*reply.NullBulkReply); !ok {
		t.Error("`get` failed, result: " + string(result.ToBytes()))
	}

	result = client.Send([][]byte{
		[]byte("DEL"),
		[]byte("arr"),
	})

	if intReply, ok := result.(*reply.IntReply); ok {
		if intReply.Code != 1 {
			t.Errorf("expected 1, but got %d", intReply.Code)
		}
	}

	result = client.Send([][]byte{
		[]byte("RPUSH"),
		[]byte("arr"),
		[]byte("1"),
		[]byte("2"),
		[]byte("3"),
	})
	if intReply, ok := result.(*reply.IntReply); ok {
		if intReply.Code != 3 {
			t.Errorf("expected 3, but got %d", intReply.Code)
		}
	}

	result = client.Send([][]byte{
		[]byte("LRANGE"),
		[]byte("arr"),
		[]byte("0"),
		[]byte("-1"),
	})
	if bulkReply, ok := result.(*reply.MultiBulkReply); ok {
		if len(bulkReply.Args) != 3 ||
			string(bulkReply.Args[0]) != "1" ||
			string(bulkReply.Args[1]) != "2" ||
			string(bulkReply.Args[2]) != "3" {
			t.Errorf("expected 1, 2, 3, but got %s, %s, %s", bulkReply.Args[0], bulkReply.Args[1], bulkReply.Args[2])
		}
	}
}
