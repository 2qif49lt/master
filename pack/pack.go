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
	// sizeHeader 协议包包头长度
	headerSize = 16
)

type cmdheader struct {
	size uint32
	cmd  uint32
	res1 uint32
	res2 uint32
}

var (
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
		cmd = phead.cmd
		body = buff[unsafe.Sizeof(*phead):]
		err = nil
		return
	}

	return 0, nil, ErrPacketBroken
}

func Pack(cmd int, body []byte) ([]byte, error) {
	buff := make([]byte, headerSize+len(body))
	phead := (*cmdheader)(unsafe.Pointer(&buff[0]))
	phead.cmd = cmd
	phead.size = uint32(headerSize + len(body))

	copy(buff[headerSize:], body)

	return buff, nil
}
