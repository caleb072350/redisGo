package reply

var (
	CRLF = "\r\n"
)

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
