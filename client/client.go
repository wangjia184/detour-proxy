package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"../comm"
	"../config"
	"../socks"

	"github.com/satori/go.uuid"
)

type ProxyClient struct {
	transport           comm.Transport
	tcpListener         net.Listener
	inaccessibleHostMap map[string]bool
	mutex               sync.RWMutex
	gfwList             *GFWList
}

func Run(httpPort uint16, socksPort uint16, uri string) error {
	this := &ProxyClient{
		inaccessibleHostMap: make(map[string]bool),
		mutex:               sync.RWMutex{},
	}

	if strings.LastIndex(uri, "?") > 0 {
		uri += "&"
	} else {
		uri += "?"
	}
	uri += "id=" + uuid.NewV4().String()

	transport, err := comm.NewWebSocketTransport(uri)
	if err != nil {
		panic(err)
	}
	this.transport = transport

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", socksPort))
	if err != nil {
		panic(err)
	}
	this.tcpListener = tcpListener

	log.Println("SOCKS server is listening on port", socksPort)
	cfg := &socks.SOCKSConf{
		Dial:        this.dial,
		HandleError: this.handleError,
	}
	go (func() {
		socks.Serve(tcpListener, cfg)
	})()

	time.AfterFunc(5*time.Second, func() { this.loadGfwList() })

	log.Println("SmartConnectTimeout =", config.GetSmartConnectTimeout())

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				this.handleTunneling(w, r)
			} else {
				this.handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Println("HTTP server is listening on port", httpPort)
	log.Fatal(server.ListenAndServe())

	return nil
}

func (this *ProxyClient) handleError(err error) {
	log.Println(err)
}

func (this *ProxyClient) dial(network, address string) (net.Conn, error) {

	if network != "tcp" {
		return nil, errors.New(fmt.Sprintf("Unsupported protocol : %v", network))
	}

	idx := strings.LastIndex(address, ":")
	if idx < 1 {
		return nil, errors.New(fmt.Sprintf("Invalid address : %v", network))
	}
	host := address[:idx]
	port, err := strconv.Atoi(address[idx+1:])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid address : %v", network))
	}

	// check host if it should not be proxied
	needProxy := (func() bool {
		this.mutex.RLock()
		defer this.mutex.RUnlock()
		return this.inaccessibleHostMap[host]
	})()
	if !needProxy {
		needProxy = this.isBlockedByGFW(host)
	}

	if !needProxy {
		smartConnectTimeout := config.GetSmartConnectTimeout()
		timeout := time.Duration(smartConnectTimeout) * time.Second
		if timeout <= 0 {
			timeout = 20 * time.Second
		}
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			if isPrivateIP(host) || smartConnectTimeout <= 0 {
				return nil, err
			}
		} else {
			return conn, nil
		}
	}
	proxyConn, err := NewProxyConnection(host, uint16(port), this.transport)
	if err != nil {
		return nil, err
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.inaccessibleHostMap[host] = true
	return proxyConn, nil
}

func isPrivateIP(ip string) bool {
	if ip == "127.0.0.1" || ip == "::1" {
		return true
	}
	IP := net.ParseIP(ip)
	if IP == nil {
		return false
	} else {
		_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
		_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
		_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
		return private24BitBlock.Contains(IP) || private20BitBlock.Contains(IP) || private16BitBlock.Contains(IP)
	}
}

func (this *ProxyClient) handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := this.dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go this.transfer(dest_conn, client_conn)
	go this.transfer(client_conn, dest_conn)
}

func (this *ProxyClient) handleHTTP(w http.ResponseWriter, req *http.Request) {

	httpTransport := &http.Transport{
		Dial: this.dial,
	}

	resp, err := httpTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	this.copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (this *ProxyClient) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (this *ProxyClient) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (this *ProxyClient) loadGfwList() {

	if len(config.GetGfwListUrl()) == 0 {
		return
	}

	httpTransport := &http.Transport{
		Dial: this.dial,
	}
	httpClient := &http.Client{Transport: httpTransport}

	// Get the data
	resp, err := httpClient.Get(config.GetGfwListUrl())
	if err != nil {
		time.AfterFunc(5*time.Second, func() { this.loadGfwList() })
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		time.AfterFunc(5*time.Second, func() { this.loadGfwList() })
		log.Println(err)
		return
	}
	base64Text := string(bodyBytes)

	this.gfwList, err = ParseRawGFWList(string(base64Text))
	if err != nil {
		time.AfterFunc(5*time.Second, func() { this.loadGfwList() })
		log.Println(err)
		return
	}
	log.Println("Loaded GFW list from", config.GetGfwListUrl())

}

func (this *ProxyClient) isBlockedByGFW(host string) bool {
	if this.gfwList != nil {
		return this.gfwList.IsBlocked(host)
	}
	return false
}
