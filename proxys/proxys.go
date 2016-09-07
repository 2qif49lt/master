package proxys

import (
	"container/list"
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

	proxys *list.List
}

type proxy struct {
	name string
	id   string

	addr *net.UDPAddr

	connid int
	alive  time.Time
}

func (ps *Proxys) CheckAlive() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for e := ps.proxys.Front(); e != nil; {
		v := e.Value.(*proxy)
		if time.Since(v.alive) > time.Second*60 {
			logrus.WithTryJson(v).Infoln("proxy is time out.")

			next := e.Next()
			ps.proxys.Remove(e)
			e = next
			continue
		}
		e = e.Next()
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
	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		v := e.Value.(*proxy)
		if v.name == req.Name {
			exist = true
			if v.id == req.Id {
				if v.addr.String() == tarAddr.String() {
					rst = msg.SUCC // re-login
				} else {
					rst = msg.RST_TRY_LATER // try later
				}
			} else {
				rst = msg.RST_NAME_EXIST
			}
			break
		}
	}

	if exist == false {
		rst = msg.SUCC
		id, _ = random.GetGuid()

		ps.proxys.PushBack(&proxy{
			req.Name,
			id,
			tarAddr,
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
		return false
	}
	return true
}

func (ps *Proxys) DoCmdAliveReq(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	req := &msg.AliveReq{}
	rsp := &msg.AliveRsp{}
	rst := msg.FAIL

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

	exist := false
	ps.lock.Lock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		v := e.Value.(*proxy)
		if v.id == req.Id && v.addr.String() == tarAddr.String() {
			exist = true
			v.alive = time.Now()
			rst = msg.SUCC
			break
		}
	}

	ps.lock.Unlock()

	if exist == false {
		rst = msg.RST_GO_LOGIN
	}

	rsp.Rst = int32(rst)

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
		return false
	}
	return true
}

func (ps *Proxys) DoCmdLogoutReq(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	req := &msg.LogoutReq{}
	rsp := &msg.LogoutRsp{}
	rst := msg.FAIL

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

	exist := false
	ps.lock.Lock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		v := e.Value.(*proxy)
		if v.id == req.Id && v.addr.String() == tarAddr.String() {
			exist = true
			rst = msg.SUCC

			ps.proxys.Remove(e)
			break
		}
	}
	ps.lock.Unlock()

	if exist == false {
		rst = msg.RST_NO_EXIST
	}

	rsp.Rst = int32(rst)

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
		return false
	}
	return true
}
