package main

import (
	"fmt"
	"os"

	"github.com/2qif49lt/master/udpcli"
)

func main() {
	cli := udpcli.New()
	//	cli.Name = "myproxycli"

	err := cli.SetSrv("127.0.0.1:8898")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cli.Run()
}
