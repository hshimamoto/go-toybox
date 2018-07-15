// simple-proxy
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./simple-proxy <listen address> <upstream proxy>
//

package main

import (
    "net/http"
    "net"
    "io"
    "os"
    "log"
    "time"
)

type Upstream struct {
    Addr string
    //
    conn *net.TCPConn
}

func (up *Upstream)handleConnect(w http.ResponseWriter,r *http.Request) {
    r.WriteProxy(up.conn)
    h, _ := w.(http.Hijacker)
    lconn, _, _ := h.Hijack()
    defer lconn.Close()

    d1 := make(chan bool)
    d2 := make(chan bool)
    go func() {
	io.Copy(up.conn, lconn)
	d1 <- true
    }()
    go func() {
	io.Copy(lconn, up.conn)
	d2 <- true
    }()
    select {
    case <-d1: go func() { <-d2 }()
    case <-d2: go func() { <-d1 }()
    }
    time.Sleep(time.Second)
}

func (up *Upstream)handleHTTP(w http.ResponseWriter,r *http.Request) {
    if r.Header.Get("Proxy-Connection") == "Keep-Alive" {
	r.Header.Set("Proxy-Connection", "close") // always close
    }
    if r.Header.Get("Connection") == "Keep-Alive" {
	r.Header.Set("Connection", "close") // always close
    }
    r.WriteProxy(up.conn)
    h, _ := w.(http.Hijacker)
    lconn, _, _ := h.Hijack()
    defer lconn.Close()

    io.Copy(lconn, up.conn)
}

func (up *Upstream)Handler(w http.ResponseWriter,r *http.Request) {
    log.Println(r.Method, r.URL)

    conn, err := net.Dial("tcp", up.Addr) // Dial to upstream
    if err != nil {
	log.Println("net.Dial:", err)
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    defer conn.Close() // don't forget close
    up.conn = conn.(*net.TCPConn)

    if r.Method == http.MethodConnect {
	up.handleConnect(w, r)
    } else {
	up.handleHTTP(w, r)
    }
}

func (up *Upstream)ListenAndServe(laddr string) {
    http.ListenAndServe(laddr, http.HandlerFunc(up.Handler))
}

func main() {
    if len(os.Args) < 3 {
	log.Fatal("Need listen address and upstream proxy")
    }
    up := &Upstream { Addr: os.Args[2] }
    up.ListenAndServe(os.Args[1])
}
