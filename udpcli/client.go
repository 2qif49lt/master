package udpcli

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/2qif49lt/logrus"
)

const (
	maxUDPMsgSize = 1024
)

type clistatus int

const (
	status_goLogin clistatus = iota
	status_keepAlive
	status_logouted
)

type Client struct {
	svrAddr  *net.UDPAddr
	conn     *net.UDPConn
	stoped   bool
	stopChan chan bool

	status clistatus

	Name string
	Id   string

	lock *sync.RWMutex // protect status name,id
}

func New() *Client {
	return &Client{
		nil,
		nil,
		false,
		make(chan bool),

		status_goLogin,
		"myproxyclient",
		"",

		&sync.RWMutex{},
	}
}

func (c *Client) SetSrv(addr string) error {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	c.svrAddr = raddr
	c.conn = conn
	return nil
}
func (c *Client) maintainStatus() {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var err error

	switch c.status {
	case status_goLogin:
		err = c.SendLoginReq()
	case status_keepAlive:
		err = c.SendAlive()
	}

	if err != nil {
		logrus.Warnln(err)
	}
}

func (c *Client) maintainRoutine() {
	time.Sleep(time.Second)
	c.maintainStatus()
	for {
		select {
		case <-c.stopChan:
			break
		case <-time.After(time.Second * 15):
			c.maintainStatus()
		}
	}
}
func (c *Client) Run() {
	go c.maintainRoutine()

	defer c.conn.Close()

	buff := [maxUDPMsgSize]byte{}
	for !c.stoped {
		data := buff[:]
		n, addr, err := c.conn.ReadFromUDP(data)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				time.Sleep(time.Second)
				logrus.Warnln("server is down.")
				continue
			} else {
				logrus.Warnln(n, addr, err)
				break
			}

		}

		err = c.handler(data[:n], addr)
		if err != nil {
			logrus.Warnln(addr, err)
		}
	}

}
func (c *Client) Stop() {
	if c.stoped {
		return
	}
	close(c.stopChan)
	c.stoped = true
}

func (c *Client) SendData(data []byte) (int, error) {
	return c.conn.Write(data)
}
