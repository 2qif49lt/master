package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/2qif49lt/logrus"
	"github.com/2qif49lt/master/pkg/random"
	"github.com/2qif49lt/master/udpcli"
	"github.com/2qif49lt/pflag"
)

var (
	proxyName   = ""
	proxyServer = ""
	proxyPort   = 0
)

func init() {
	pflag.StringVarP(&proxyName, "name", "n", "", `your proxy client's name`)
	pflag.StringVarP(&proxyServer, "server", "s", "", "proxy server address")
	pflag.IntVarP(&proxyPort, "port", "p", 8898, "proxy server port")
}

func main() {
	pflag.Parse()

	if proxyName == "" {
		proxyName, _ = random.GetGuid()
	}

	if strings.HasPrefix(proxyName, "show") == false {
		proxyName = "show" + proxyName
	}

	addrs, err := net.LookupHost(proxyServer)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	fmt.Println(addrs, err)
	logrus.WithFields(
		logrus.Fields{
			"name":   proxyName,
			"server": addrs[0],
			"visit":  fmt.Sprintf("http://%s.%s", proxyName, proxyServer),
		}).Infoln("proxy client start")

	srvaddr := fmt.Sprintf("%s:%d", addrs[0], proxyPort)

	cli := udpcli.New()
	cli.Name = proxyName

	err = cli.SetSrv(srvaddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cli.Run()
}
