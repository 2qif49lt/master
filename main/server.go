package main

import (
	"fmt"
	"github.com/2qif49lt/master/httpsrv"
	"github.com/2qif49lt/master/proxys"
	"github.com/2qif49lt/master/udpsrv"
)

func main() {
	srv := udpsrv.New()
	ps := proxys.New()
	ps.Udpsender = srv
	srv.SetHander(&udpsrv.MsgHandler{ps})
	go httpsrv.RunOn(":8080", ps, ":7898")
	err := srv.RunOn(":8898")
	fmt.Println(err)

}
