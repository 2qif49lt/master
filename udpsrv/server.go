package udpsrv

import (
	"fmt"
	"net"
	"time"
)

const (
	// 默认消息队列长度
	defaultMaxQueueSize = 10 * 1024
	// 默认客户端超时时间，秒
	defaultMaxTimeOut = 60
	// 默认单个包最大长度
	defaultMaxPacketSize = 10 * 1024
)

type Srv struct {
	MaxQueueSize int // 消息队列长度
	MaxTimeOut   int // 超时时长

	stopchan chan bool // 服务停止
	lconn    *net.UDPConn
	stoped   bool

	sendqueue chan *inQueueMsg
	recvqueue chan *inQueueMsg

	handler Handler
}

type inQueueMsg struct {
	putTime    time.Time
	remoteAddr *net.UDPAddr
	data       []byte
}

func (srv *Srv) Send(msg []byte, tarAddr *net.UDPAddr) error {
	if msg == nil {
		return fmt.Errorf("msg is nil")
	}
	packet := &inQueueMsg{}

	packet.putTime = time.Now()
	packet.remoteAddr = tarAddr
	packet.data = msg

	srv.sendqueue <- packet

	return nil
}

func New() *Srv {
	return &Srv{
		defaultMaxQueueSize,
		defaultMaxTimeOut,

		make(chan bool),
		nil,
		false,

		make(chan *inQueueMsg, defaultMaxQueueSize),
		make(chan *inQueueMsg, defaultMaxQueueSize),

		&defHandler{},
	}
}

func (s *Srv) run() {
	defer s.lconn.Close()
	for !s.stoped {
		data := make([]byte, defaultMaxPacketSize)

		n, remoteAddr, err := s.lconn.ReadFromUDP(data)
		if err == nil {
			msg := &inQueueMsg{}

			msg.putTime = time.Now()
			msg.remoteAddr = remoteAddr
			msg.data = data[:n]

			s.recvqueue <- msg
		}
	}
}

// 重新启动服务 应该重新的New,然后RunOn
func (s *Srv) RunOn(addr string) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	lconn, lerr := net.ListenUDP("udp", laddr)
	if lerr != nil {
		return lerr
	}
	s.lconn = lconn

	go s.handlemsg()
	go s.sendmsg()

	s.run()

	return nil
}

func (s *Srv) Stop() {
	if s.stoped {
		return
	}

	close(s.stopchan)
	s.stoped = true
}

func (s *Srv) handlemsg() {
	for {
		select {
		case <-s.stopchan:
			break
		case inmsg, ok := <-s.recvqueue:
			if ok && s.handler != nil {
				s.handler.Handle(inmsg, s)
			}
		}
	}
}
func (s *Srv) sendmsg() {
	for {
		select {
		case <-s.stopchan:
			break
		case inmsg, ok := <-s.sendqueue:
			if ok {
				s.lconn.WriteToUDP(inmsg.data, inmsg.remoteAddr)
			}
		}
	}
}
