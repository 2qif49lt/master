package proxys

import (
	"net"
	"sync"
	"time"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/msg"
	"github.com/2qif49lt/master/pack"
	"github.com/2qif49lt/master/pkg/random"
	"github.com/golang/protobuf/proto"
)

// Sender is generic interface to send msg to server
type Sender interface {
	Send(msg []byte, tarAddr *net.UDPAddr) error
}

// Proxys represent proxy run in the client host who behind a nat
type Proxys struct {
	lock sync.Mutex

	proxys []proxy
}

type proxy struct {
	name string
	id   string

	connid int
	alive  time.Time
}

func (ps *Proxys) CheckAlive() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for k, v := range ps.proxys {
		if time.Since(v.alive) > time.Second*60 {
			logrus.WithTryJson(v).Infoln("proxy is time out.")
			//  ps.proxys = append(ps.proxys[:k],ps.proxys[k+1:])
			ps.proxys = ps.proxys[:k+copy(ps.proxys[k:], ps.proxys[k+1:])]
		}
	}
}
func (ps *Proxys) DoCmdLoginReq(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	req := &msg.LoginReq{}
	rsp := &msg.LoginRsp{}
	rst := msg.FAIL
	id := ""

	err := proto.Unmarshal(body, req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"time":     recvTime.Unix(),
			"data len": len(body),
			"error":    err.Error(),
		}).Warnln("proto unmarshal fail.")
		return false
	}

	ps.lock.Lock()
	exist := false
	for _, v := range ps.proxys {
		if v.name == req.Name {
			exist = true
			if v.id == req.Id {
				rst = msg.SUCC
			} else {
				rst = msg.RST_NAME_EXIST
			}
			break
		}
	}

	if exist == false {
		rst = msg.SUCC
		id, _ = random.GetGuid()

		ps.proxys = append(ps.proxys, proxy{
			req.Name,
			id,
			0,
			time.Now(),
		})
	}
	ps.lock.Unlock()

	rsp.Rst = int32(rst)
	rsp.Id = id

	bodydata, marshalerr := proto.Marshal(rsp)
	if marshalerr != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":   cmd,
			"ip":    tarAddr.String(),
			"error": marshalerr.Error(),
		}).Warnln("proto marshal fail.")
		return false
	}

	senddata, packerr := pack.Pack(msg.CMD_LOGIN_RSP, bodydata)
	if packerr != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(bodydata),
			"error":    packerr.Error(),
		}).Warnln("Pack fail.")
		return false
	}
	err = sender.Send(senddata, tarAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(senddata),
			"error":    err.Error(),
		}).Warnln("Send fail.")
	}
	return true
}

func (srvs *Proxys) DoCmdAliveReq(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	return false
}

func (srvs *Proxys) DoCmdLogoutReq(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	return false
}
