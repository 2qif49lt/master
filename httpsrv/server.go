package httpsrv

import (
	//	"bytes"
	"context"
	"fmt"
	//	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
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
func checkSubDomain(host string) string {
	arr := strings.Split(host, ".")
	if len(arr) != 3 {
		return ""
	}
	return arr[0]
}
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subDomain := checkSubDomain(r.Host)
	if subDomain == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	rProxy := &httputil.ReverseProxy{}
	rProxy.Director = func(r *http.Request) {
		r.URL.Scheme = "http"
		r.Host = "127.0.0.1"
		r.URL.Host = r.Host
	}

	connid, e := h.ps.SendNewConnPushReqWithAutoRandId(subDomain, h.tcpAddr, 80)
	if e != nil {
		fmt.Println(e, connid)
		w.Write([]byte(e.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rProxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return h.waitForReverseConnection(ctx, subDomain, connid)
		},
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	rProxy.ServeHTTP(w, r)

}

func (h *handler) ServeHTTPPath(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.RequestURI, "/favicon.ico") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println("req uri", r.RequestURI)
	if r.Referer() != "" {
		u, e := url.Parse(r.Referer())
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(e.Error()))
			return
		}
		fmt.Println("refer host", u.Host, "refer path", u.Path, "rhost", r.Host, "r refer", r.Referer())
		if u.Host == r.Host {

			n, p, _, e := checkUri(u.Path)
			if e != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(fmt.Sprintf("refer %s", e.Error())))
				return
			}
			if p == 80 {
				r.URL.Path = fmt.Sprintf("/%s%s", n, r.URL.Path)
			} else {
				r.URL.Path = fmt.Sprintf("/%s/%d%s", n, p, r.URL.Path)
			}
			n1, p1, _, e := checkUri(r.URL.Path)
			if n1 != n || p1 != p {
				w.Header().Set("Location", r.URL.String())
				w.WriteHeader(http.StatusMovedPermanently)
				return
			}
			r.Header.Set("Referer", "")

			fmt.Println("after r.URL.Path", r.URL.Path)
		}
	}

	fmt.Printf("%#v\n", r.URL)

	rProxy := &ReverseProxy{}

	r.URL.Scheme = "http"
	name, port, s, e := checkUri(r.URL.Path)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	if s == "" {
		s = "/"
	}

	rProxy.Director = func(r *http.Request) {
		// already done outside
		fmt.Println("Director", r.RequestURI, r.URL)
	}

	rProxy.DoLocation = func(old string) (string, bool) {
		return "", false
		ret := ""
		if port == 80 {
			ret = fmt.Sprintf("/%s/%s", name, old)
		} else {
			ret = fmt.Sprintf("/%s/%d/%s", name, port, old)
		}
		ret = path.Clean(ret)
		fmt.Println("dolocation", old, ret)
		return ret, true
	}

	referer := r.Header.Get("Referer")
	if referer != "" {

	}
	r.Host = "127.0.0.1"
	if port != 80 {
		r.Host = fmt.Sprintf("127.0.0.1:%d", port)
	}
	r.URL.Host = r.Host
	r.URL.Path = s

	fmt.Printf("%#v %s\n", r.URL, r.URL.String())

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
