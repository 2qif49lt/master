package server

import (
	"fmt"
	"net"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/msg"
	"github.com/2qif49lt/master/pack"
	"github.com/2qif49lt/master/proxys"
	"github.com/golang/protobuf/proto"
)

type Srv struct {
	ps *proxys.Proxys
}

func New(ps *proxys.Proxys) *Srv {
	return &Srv{ps}
}

func (srv *Srv) RunOn(addr string) error {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	var tempDelay time.Duration
	defer lis.Close()
	for {
		rw, e := lis.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Warnf("http: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := srv.newConn(rw)
		go c.serve(ctx)
	}
}

type Conn struct {
	srv    *Srv
	conn   net.Conn
	id     string
	connid string
}

func (srv *Srv) newConn(rw net.Conn) *Conn {
	c := &Conn{srv, rw}
	return c
}

func (c *Conn) serve() {

	for {

	}
}

func (c *Conn) readAndServeNewConnPushRsp() error {
	needSize := pack.HeadSize()
	buf := make([]byte, needSize)
	nRead := 0

	for nRead < needSize {
		nPer, err := c.conn.Read(buf[nRead:])
		if err != nil {
			return err
		}
		nRead += nPer
	}

	cmd, bodyLen, err := pack.UnpackHead(buf)
	if err != nil {
		return err
	}

	if cmd != msg.CMD_NEW_CONN_PUSH_RSP {
		return fmt.Errorf("unexpected cmd: %d", cmd)
	}

	needSize = bodyLen
	buf = make([]byte, needSize)
	nRead = 0

	for nRead < needSize {
		nPer, err := c.conn.Read(buf[nRead:])
		if err != nil {
			return err
		}
		nRead += nPer
	}

	rsp := &msg.NewConnPushRsp{}

	err = proto.Unmarshal(buf, rsp)
	if err != nil {
		return err
	}

	err = c.srv.ps.CheckNewConnPushRsp(rsp.Id, rsp.Connid)
	if err != nil {
		return err
	}
	c.id = rsp.Id
	c.connid = rsp.Connid

	return nil
}
