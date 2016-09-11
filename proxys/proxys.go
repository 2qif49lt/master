package proxys

import (
	"container/list"
	"fmt"
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
	Udpsender Sender

	lock   *sync.Mutex
	proxys *list.List
}

func New() *Proxys {
	p := &Proxys{
		nil,

		&sync.Mutex{},
		list.New(),
	}
	return p
}

func (ps *Proxys) SetSender(sender Sender) {
	ps.Udpsender = sender
}

type proxy struct {
	name string
	id   string

	addr *net.UDPAddr

	conns map[string]*Proxyconn
	alive time.Time
}

type Proxyconn struct {
	reqTime time.Time
	cliconn chan *net.TCPConn
}

func (ps *Proxys) PushProxyReverseConnChan(id, connid string, conn *net.TCPConn) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		pr := e.Value.(*proxy)

		if pr.id == id {
			if pc, ok := pr.conns[connid]; ok {
				if time.Since(pc.reqTime) > time.Second*15 {
					return fmt.Errorf("PushProxyReverseConnChan connection wait timeout,id: %s;connid: %s", id, connid)
				}
				pc.cliconn <- conn
				return nil
			} else {
				return fmt.Errorf("PushProxyReverseConnChan connection is invalid,id: %s;connid: %s", id, connid)
			}
		}
	}
	return fmt.Errorf("PushProxyReverseConnChan proxy not found,id: %s", id)
}
func (ps *Proxys) GetProxyReverseConnChan(name, connid string) (chan *net.TCPConn, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		pr := e.Value.(*proxy)

		if pr.name == name {
			if pc, ok := pr.conns[connid]; ok {
				if time.Since(pc.reqTime) > time.Second*15 {
					return nil, fmt.Errorf("connection wait timeout,name: %s;connid: %s", name, connid)
				}
				//	delete(pr.conns, connid)
				return pc.cliconn, nil
			} else {
				return nil, fmt.Errorf("connection is invalid,name: %s;connid: %s", name, connid)
			}
		}
	}
	return nil, fmt.Errorf("proxy not found,name: %s", name)
}
func (ps *Proxys) CheckNewConnPushRsp(id, connid string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		pr := e.Value.(*proxy)
		if pr.id == id {
			if pc, ok := pr.conns[connid]; ok {
				if time.Since(pc.reqTime) <= time.Second*15 {
					return nil
				}
			}

		}
	}
	return fmt.Errorf("CheckNewConnPushRsp not found")
}
func (ps *Proxys) SendNewConnPushReqWithAutoRandId(name, srvAddr string, port int) (string, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for e := ps.proxys.Front(); e != nil; e = e.Next() {
		v := e.Value.(*proxy)
		if v.name == name {
			randid, _ := random.GetGuid()
			v.conns[randid] = &Proxyconn{time.Now(), make(chan *net.TCPConn, 1)}
			return randid, v.sendNewConnPushReq(randid, srvAddr, port, ps.Udpsender)
		}
	}
	return "", fmt.Errorf("SendNewConnPushReqWithAutoRandId not found,name:%s,srvaddr:%s", name, srvAddr)
}
func (p *proxy) sendNewConnPushReq(connid, srvAddr string, port int, sender Sender) error {
	req := &msg.NewConnPushReq{}
	req.Connid = connid
	req.Srvaddr = srvAddr
	req.Locport = int32(port)

	bodydata, marshalerr := proto.Marshal(req)
	if marshalerr != nil {
		return marshalerr
	}

	senddata, packerr := pack.Pack(msg.CMD_NEW_CONN_PUSH_REQ, bodydata)
	if packerr != nil {
		return packerr
	}
	err := sender.Send(senddata, p.addr)
	if err != nil {
		return err
	}
	logrus.WithTryJson(req).Infoln("sendNewConnPushReq")
	return nil
}
func (ps *Proxys) CheckAlive(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			break
		case <-time.After(time.Second * 15):
			ps.lock.Lock()
			for e := ps.proxys.Front(); e != nil; {
				v := e.Value.(*proxy)
				if time.Since(v.alive) > time.Second*60 {
					logrus.WithFields(logrus.Fields{
						"name": v.name,
						"id":   v.id,
						"addr": v.addr.String(),
					}).Warnln("proxy is time out.")
					next := e.Next()
					ps.proxys.Remove(e)
					e = next
					continue
				}
				e = e.Next()
			}
			ps.lock.Unlock()
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

	logrus.WithTryJson(req).Infoln("DoCmdLoginReq")

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
			make(map[string]*Proxyconn),
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
	logrus.WithTryJson(rsp).Infoln("DoCmdLoginReq")

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
	//	logrus.WithTryJson(req).Infoln("DoCmdAliveReq")

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
	senddata, packerr := pack.Pack(msg.CMD_ALIVE_RSP, bodydata)
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
	//	logrus.WithTryJson(rsp).Infoln("DoCmdAliveReq")
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

/*
func (ps *Proxys) DoCmdNewConnPushRsp(cmd int, body []byte, sender Sender, recvTime time.Time, tarAddr *net.UDPAddr) bool {
	rsp := &msg.NewConnPushRsp{}
	rst := msg.FAIL

	err := proto.Unmarshal(body, rsp)
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
*/
