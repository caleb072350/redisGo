package reply

import (
	"redisGo/interface/redis"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

/* ---- Bulk Reply ---- */
type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

/* ---- Multi Bulk Reply ---- */
type MultiBulkReply struct {
	Args [][]byte
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	res := "*" + strconv.Itoa(argLen) + CRLF
	for _, arg := range r.Args {
		if arg == nil {
			res += "$-1" + CRLF
		} else {
			res += "$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF
		}
	}
	return []byte(res)
}

/* ---- Error Reply ----- */
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

type StandardErrReply struct {
	Status string
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

func IsErrorReply(reply redis.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

/* ---- Int Reply ----- */
type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

/* ---- Status Reply ---- */

type StatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + "\r\n")
}
