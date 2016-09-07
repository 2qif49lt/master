package main

import (
	"fmt"
	"github.com/2qif49lt/master/proxys"
	"github.com/2qif49lt/master/udpsrv"
)

func main() {
	srv := udpsrv.New()
	srv.SetHander(&udpsrv.MsgHandler{proxys.New()})
	err := srv.RunOn(":8898")
	fmt.Println(err)

}
