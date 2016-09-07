package udpsrv

import (
	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/msg"
	"github.com/2qif49lt/master/pack"
	"github.com/2qif49lt/master/proxys"
)

type msgHandler struct {
	clients *proxys.Proxys
}

func (hder *msgHandler) Handle(inmsg *inQueueMsg, sender Sender) bool {
	cmd, body, err := pack.Unpack(inmsg.data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ip":       inmsg.remoteAddr.String(),
			"time":     inmsg.putTime,
			"data len": len(inmsg.data),
			"error":    err.Error(),
		}).Warnln("unpack fail")
		return false
	}
	switch cmd {
	case msg.CMD_LOGIN_REQ:
		return hder.clients.DoCmdLoginReq(cmd, body, sender, inmsg.putTime, inmsg.remoteAddr)
	case msg.CMD_ALIVE_REQ:
		return hder.clients.DoCmdAliveReq(cmd, body, sender, inmsg.putTime, inmsg.remoteAddr)
	case msg.CMD_LOGOUT_REQ:
		return hder.clients.DoCmdLogoutReq(cmd, body, sender, inmsg.putTime, inmsg.remoteAddr)
	}
	return false
}
