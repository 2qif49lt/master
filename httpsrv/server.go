package httpsrv

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/2qif49lt/master/proxys"
)

const (
	// expect [/{name}/{port} {name} /{port} {port}]
	targetCliExp = `^/(\w{3,})(/(\d+)/)?`
)

//var targetCliReg = regexp.MustCompile(targetCliExp)

func checkUri(uri string) (name string, port int, subUri string, err error) {
	arr := strings.Split(uri, "/")
	if len(arr) > 0 {
		arr = arr[1:]
	}
	if len(arr) > 0 && arr[len(arr)-1] == "" {
		arr = arr[:len(arr)-1]
	}
	if len(arr) == 0 {
		err = fmt.Errorf(`path wrong`)
		return
	}

	name = arr[0]
	subUri = strings.TrimPrefix(uri, "/"+name)
	if len(arr) > 1 {
		port, err = strconv.Atoi(arr[1])
		if err == nil {
			subUri = strings.TrimPrefix(subUri, "/"+arr[1])
			return
		} else {
			err = nil
		}
	}
	port = 80

	return
}

type handler struct {
	ps      *proxys.Proxys
	tcpAddr string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n, p, s, e := checkUri(r.RequestURI)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	if s == "" {
		s = "/"
	}
	fmt.Println(n, p, s, e)

	buf := bytes.NewBuffer(nil)
	r.Host = "127.0.0.1"
	if p != 80 {
		r.Host = fmt.Sprintf("127.0.0.1:%d", p)
	}
	r.URL.Path = s
	e = r.Write(buf)
	if e != nil {
		fmt.Println(e)
		return
	}

	srcdata, e := ioutil.ReadAll(buf)
	if e != nil {
		fmt.Println(e)
		return
	}

	w.Write(srcdata)
	h.ps.SendNewConnPushReqWithAutoRandId(n, h.tcpAddr)
}

func RunOn(addr string, ps *proxys.Proxys, tcpAddr string) error {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	h := &handler{ps, tcpAddr}
	return http.Serve(lis, h)
}
