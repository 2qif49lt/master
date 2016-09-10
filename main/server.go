package main

import (
	"fmt"
	"github.com/2qif49lt/master/httpsrv"
	"github.com/2qif49lt/master/proxys"
	tcpsrv "github.com/2qif49lt/master/tcp/server"
	"github.com/2qif49lt/master/udpsrv"
)

func main() {
	udpSrv := udpsrv.New()
	ps := proxys.New()
	ps.Udpsender = udpSrv
	udpSrv.SetHander(&udpsrv.MsgHandler{ps})
	go httpsrv.RunOn(":8080", ps, "127.0.0.1:7898")
	tcpSrv := tcpsrv.New(ps)
	go tcpSrv.RunOn(":7898")
	err := udpSrv.RunOn(":8898")
	fmt.Println(err)

}
