package pack

import (
	"errors"
	"unsafe"
)

// 拆包

const (
	// sizeFieldSize 协议包包头的size字段长度
	sizeFieldSize = 4
	// sizeFieldPos size字段在包头的位置
	sizeFieldPos = 0
)

var (
	// sizeHeader 协议包包头长度
	headerSize = int(unsafe.Sizeof(cmdheader{}))
)

func HeadSize() int {
	return headerSize
}

func UnpackHead(buff []byte) (cmd, bodyLen int, err error) {
	if len(buff) != headerSize {
		err = ErrHeaderWrong
		return
	}
	phead := (*cmdheader)(unsafe.Pointer(&buff[0]))
	cmd = int(phead.cmd)
	bodyLen = int(phead.size) - headerSize
	err = nil
	return
}

type cmdheader struct {
	size uint32
	cmd  uint32
	res1 uint32
	res2 uint32
}

var (
	ErrHeaderWrong  = errors.New("头解析出错")
	ErrPacketBroken = errors.New("包不完整")
	ErrSizeInvalid  = errors.New("头部大小错误") // 头部大小错误
)

func Unpack(buff []byte) (cmd int, body []byte, err error) {
	if buff == nil || len(buff) < sizeFieldSize+sizeFieldPos {
		return 0, nil, ErrPacketBroken
	}
	msgsize := *(*uint32)(unsafe.Pointer(&buff[sizeFieldPos]))

	if msgsize < uint32(headerSize) {
		return 0, nil, ErrSizeInvalid
	}
	if len(buff) == int(msgsize) {
		phead := (*cmdheader)(unsafe.Pointer(&buff[0]))
		cmd = int(phead.cmd)
		body = buff[unsafe.Sizeof(*phead):]
		err = nil
		return
	}

	return 0, nil, ErrPacketBroken
}

func Pack(cmd int, body []byte) ([]byte, error) {
	buff := make([]byte, headerSize+len(body))
	phead := (*cmdheader)(unsafe.Pointer(&buff[0]))
	phead.cmd = uint32(cmd)
	phead.size = uint32(headerSize + len(body))

	copy(buff[headerSize:], body)

	return buff, nil
}
