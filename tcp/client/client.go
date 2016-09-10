package client

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/msg"
	"github.com/2qif49lt/master/pack"
	"github.com/golang/protobuf/proto"
)

type ClientConn struct {
	Id        string
	ConnId    string
	LocalPort int
	SrvAddr   string

	srvconn *net.TCPConn
	locconn *net.TCPConn

	wg sync.WaitGroup
}

func (conn *ClientConn) DoProxy() {
	locaddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", conn.LocalPort))
	if err != nil {
		logrus.Errorln(err)
		return
	}

	conn.locconn, err = net.DialTCP("tcp", nil, locaddr)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	srvaddr, err := net.ResolveTCPAddr("tcp", conn.SrvAddr)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	conn.srvconn, err = net.DialTCP("tcp", nil, srvaddr)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	err = conn.SendNewConnPushRsp()
	if err != nil {
		logrus.Errorln(err)
		return
	}

	conn.wg.Add(1)
	go func() {
		io.Copy(conn.locconn, conn.srvconn)
		conn.locconn.Close()
		conn.wg.Done()
	}()
	conn.wg.Add(1)
	go func() {
		io.Copy(conn.srvconn, conn.locconn)
		conn.srvconn.Close()
		conn.wg.Done()
	}()
	conn.wg.Wait()
}

func (conn *ClientConn) SendNewConnPushRsp() error {
	rsp := &msg.NewConnPushRsp{}
	rsp.Id = conn.Id
	rsp.Connid = conn.ConnId
	bodydata, marshalerr := proto.Marshal(rsp)
	if marshalerr != nil {
		return marshalerr
	}
	senddata, packerr := pack.Pack(msg.CMD_NEW_CONN_PUSH_RSP, bodydata)
	if packerr != nil {
		return packerr
	}

	totalSize := len(senddata)
	nSend := 0

	for nSend < totalSize {
		nPer, err := conn.srvconn.Write(senddata[nSend:])
		if err != nil {
			return err
		}
		nSend += nPer
	}

	return nil
}
