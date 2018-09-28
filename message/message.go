package l2t

type (
	msgType int
)

const (
	//	l2tV1   = 1
	//	l2tPort = 2228

	requestDst = msgType(1)
	requestSrc = msgType(2)
	replyDst   = msgType(3)
	replySrc   = msgType(4)
)

//type l2tMsg struct {
//	l2tMsgType byte
//	l2tVer     byte
//	attrs      []attr
//}

var (
	msgTypeString = map[msgType]string{
		requestDst: "L2T_REQUEST_DST",
		requestSrc: "L2T_REQUEST_SRC",
		replyDst:   "L2T_REPLY_DST",
		replySrc:   "L2T_REPLY_SRC",
	}
)
