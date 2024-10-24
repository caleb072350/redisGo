package reply

import "strconv"

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

/* ---- Error Reply ---- */
type ErrReply struct {
	Status string
}

func (r *ErrReply) ToBytes() []byte {
	return []byte(("-") + r.Status + CRLF)
}

func MakeErrReply(status string) *ErrReply {
	return &ErrReply{
		Status: status,
	}
}

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
