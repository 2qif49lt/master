package udpcli

import (
	"fmt"
	"net"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/msg"
	"github.com/2qif49lt/master/pack"
	tcpclient "github.com/2qif49lt/master/tcp/client"
	"github.com/golang/protobuf/proto"
)

func (c *Client) handler(data []byte, tarAddr *net.UDPAddr) error {
	cmd, body, err := pack.Unpack(data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ip":       tarAddr.String(),
			"data len": len(data),
			"error":    err.Error(),
		}).Warnln("unpack fail")
		return err
	}
	switch cmd {
	case msg.CMD_LOGIN_RSP:
		return c.handleLoginRsp(cmd, body, tarAddr)
	case msg.CMD_ALIVE_RSP:
		return c.handleAliveRsp(cmd, body, tarAddr)
	case msg.CMD_LOGOUT_RSP:
		return c.handleLogoutRsp(cmd, body, tarAddr)
	case msg.CMD_NEW_CONN_PUSH_REQ:
		return c.handleNewConnPushReq(cmd, body, tarAddr)

	default:
		return fmt.Errorf("cmd: %d donot support", cmd)
	}
	return nil
}

func (c *Client) SendLoginReq() error {
	cmd := msg.CMD_LOGIN_REQ
	req := &msg.LoginReq{}

	req.Name = c.Name
	req.Id = c.Id

	body, marshalerr := proto.Marshal(req)
	if marshalerr != nil {
		return marshalerr
	}

	data, packerr := pack.Pack(cmd, body)
	if packerr != nil {
		return packerr
	}

	_, err := c.SendData(data)
	logrus.WithTryJson(req).Infoln("SendLoginReq")

	return err
}
func (c *Client) handleLoginRsp(cmd int, body []byte, tarAddr *net.UDPAddr) error {

	rsp := &msg.LoginRsp{}

	err := proto.Unmarshal(body, rsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(body),
			"error":    err.Error(),
		}).Warnln("proto unmarshal fail.")
		return err
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	switch rsp.Rst {
	case msg.SUCC:
		c.status = status_keepAlive
		c.Id = rsp.Id
	default:
		c.status = status_goLogin
	}

	logrus.WithTryJson(rsp).Infoln("handleLoginRsp")

	return nil
}

func (c *Client) SendAlive() error {

	cmd := msg.CMD_ALIVE_REQ
	req := &msg.AliveReq{}

	req.Id = c.Id

	body, marshalerr := proto.Marshal(req)
	if marshalerr != nil {
		return marshalerr
	}

	data, packerr := pack.Pack(cmd, body)
	if packerr != nil {
		return packerr
	}

	_, err := c.SendData(data)

	//	logrus.WithTryJson(req).Infoln("SendAlive")

	return err
}
func (c *Client) handleAliveRsp(cmd int, body []byte, tarAddr *net.UDPAddr) error {

	rsp := &msg.AliveRsp{}

	err := proto.Unmarshal(body, rsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(body),
			"error":    err.Error(),
		}).Warnln("proto unmarshal fail.")
		return err
	}
	if rsp.Rst != msg.SUCC {
		c.lock.Lock()
		c.status = status_goLogin
		c.lock.Unlock()
	}

	//	logrus.WithTryJson(rsp).Infoln("handleAliveRsp")

	return nil
}

func (c *Client) SendLogout() error {
	cmd := msg.CMD_LOGOUT_REQ
	req := &msg.LogoutReq{}

	req.Id = c.Id

	body, marshalerr := proto.Marshal(req)
	if marshalerr != nil {
		return marshalerr
	}

	data, packerr := pack.Pack(cmd, body)
	if packerr != nil {
		return packerr
	}

	_, err := c.SendData(data)
	return err
}
func (c *Client) handleLogoutRsp(cmd int, body []byte, tarAddr *net.UDPAddr) error {
	rsp := &msg.LogoutRsp{}

	err := proto.Unmarshal(body, rsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(body),
			"error":    err.Error(),
		}).Warnln("proto unmarshal fail.")
		return err
	}

	if rsp.Rst == msg.SUCC {
		c.lock.Lock()
		c.status = status_logouted
		c.lock.Unlock()
	}

	return nil
}

func (c *Client) handleNewConnPushReq(cmd int, body []byte, tarAddr *net.UDPAddr) error {
	req := &msg.NewConnPushReq{}

	err := proto.Unmarshal(body, req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      cmd,
			"ip":       tarAddr.String(),
			"data len": len(body),
			"error":    err.Error(),
		}).Warnln("proto unmarshal fail.")
		return err
	}

	logrus.WithTryJson(req).Infoln("handleNewConnPushReq")

	tcpcli := &tcpclient.ClientConn{}
	tcpcli.Id = c.Id
	tcpcli.ConnId = req.Connid
	tcpcli.LocalPort = int(req.Locport)
	tcpcli.SrvAddr = req.Srvaddr
	if len(req.Srvaddr) > 0 && req.Srvaddr[0] == ':' {
		tcpcli.SrvAddr = fmt.Sprintf("%s%s", c.svrAddr.IP.To4().String(), req.Srvaddr)
	}

	go tcpcli.DoProxy()
	return nil
}
