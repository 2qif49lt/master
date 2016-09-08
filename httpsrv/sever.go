package httpsrv

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var counter = 1

const (
	// expect [/{name}/{port} {name} /{port} {port}]
	targetCliExp = `^/([a-z0-9A-Z]{3,32})(/(\d{2,5}))?`
)

var targetCliReg = regexp.MustCompile(targetCliExp)

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	matchArr := targetCliReg.FindStringSubmatch(r.RequestURI)
	if matchArr == nil || len(matchArr) != 4 || matchArr[0] == "" || matchArr[1] == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("path wrong"))
		return
	}
	fmt.Println(r.RequestURI, len(matchArr), matchArr)

	name := matchArr[1]
	port := 80
	uri := strings.Replace(r.RequestURI, matchArr[0], "", 1)
	if uri == "" || uri[0] != '/' {
		uri = "/" + uri
	}

	if matchArr[3] != "" {
		port, _ = strconv.Atoi(matchArr[3])
	}

	w.Write([]byte(fmt.Sprintf(`hello %d 
        name:%s
        port:%d
        new uri:%s.`, counter, name, port, uri)))
	counter++
}

func Run(addr string) error {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	return http.Serve(lis, http.HandlerFunc(proxyHandler))
}
