package udpsrv

import (
//"errors"
)

type msgHandler struct {
}

func (def *msgHandler) Handle(inmsg *inQueueMsg, sender Sender) bool {
	return false
}
