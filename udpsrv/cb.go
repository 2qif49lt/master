package udpsrv

import (
	"net"
)

// Sender is generic interface to send msg to server
type Sender interface {
	Send(msg []byte, tarAddr *net.UDPAddr) error
}

type Handler interface {
	Handle(inmsg *inQueueMsg, sender Sender) bool
}

// SetHander should be call before run
func (srv *Srv) SetHander(handle Handler) {
	srv.handler = handle
}

type defHandler struct {
}

func (def *defHandler) Handle(inmsg *inQueueMsg, sender Sender) bool {
	return sender.Send(inmsg.data, inmsg.remoteAddr) == nil
}
