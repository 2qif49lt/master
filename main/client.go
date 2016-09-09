package main

import (
	"fmt"
	"os"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/pkg/random"
	"github.com/2qif49lt/master/udpcli"
	"github.com/2qif49lt/pflag"
)

var (
	proxyName = ""
)

func init() {
	pflag.StringVarP(&proxyName, "name", "n", "", "your proxy client's name, it will be a random name if not set")
}

func main() {
	pflag.Parse()

	if proxyName == "" {
		proxyName, _ = random.GetGuid()
	}
	logrus.WithField("name", proxyName).Infoln("proxy client start")

	cli := udpcli.New()
	cli.Name = proxyName

	err := cli.SetSrv("127.0.0.1:8898")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cli.Run()
}
