package httpsrv

import (
	//	"bytes"
	"context"
	"fmt"
	//	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	//	"github.com/2qif49lt/logrus"
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

func (h *handler) waitForReverseConnection(ctx context.Context, name, connid string) (*net.TCPConn, error) {
	cliChan, err := h.ps.GetProxyReverseConnChan(name, connid)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("wait cancel")
	case <-time.After(time.Second * 30):
		return nil, fmt.Errorf("wait timeout")
	case cliConn, ok := <-cliChan:
		if ok == false {
			return nil, fmt.Errorf("proxy client wait channel closed")
		}
		return cliConn, nil
	}
}
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.RequestURI, "/favicon.ico") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	rProxy := &httputil.ReverseProxy{}
	rProxy.Director = func(r *http.Request) {
		// already done outside
	}
	r.URL.Scheme = "http"
	name, port, s, e := checkUri(r.RequestURI)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	if s == "" {
		s = "/"
	}
	fmt.Println(name, port, s, e)

	r.Host = "127.0.0.1"
	if port != 80 {
		r.Host = fmt.Sprintf("127.0.0.1:%d", port)
	}
	r.URL.Host = r.Host
	r.URL.Path = s

	connid, e := h.ps.SendNewConnPushReqWithAutoRandId(name, h.tcpAddr, port)
	if e != nil {
		fmt.Println(e, connid)
		w.Write([]byte(e.Error()))
		return
	}

	rProxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return h.waitForReverseConnection(ctx, name, connid)
		},
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	rProxy.ServeHTTP(w, r)

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
