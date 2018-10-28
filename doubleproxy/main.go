// doubleproxy
//
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "time"
)

func Transfer(lconn, rconn net.Conn) {
    d1 := make(chan bool)
    d2 := make(chan bool)
    go func() {
	io.Copy(rconn, lconn)
	d1 <- true
    }()
    go func() {
	io.Copy(lconn, rconn)
	d2 <- true
    }()
    select {
    case <-d1: go func() { <-d2 }()
    case <-d2: go func() { <-d1 }()
    }
    time.Sleep(time.Second)
}

type Proxyes struct {
    proxy1, proxy2 string
}

func (p *Proxyes)handleConnect(w http.ResponseWriter,r *http.Request) {
    port := r.URL.Port()

    rconn, err := net.Dial("tcp", p.proxy1) // Dial to upstream
    if err != nil {
	log.Println("net.Dial:", err)
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    defer rconn.Close() // don't forget close

    if port != "443" {
	// start hijacking
	h, _ := w.(http.Hijacker)
	lconn, _, _ := h.Hijack()
	defer lconn.Close()

	r.WriteProxy(rconn)
	Transfer(lconn, rconn)
	return
    }

    rconn.Write([]byte("CONNECT " + p.proxy2 + " HTTP/1.1\r\n\r\n"))
    buf := make([]byte, 256)
    n, err := rconn.Read(buf)
    if n == 0 {
	return
    }
    // must receive HTTP/1.1 200 Established, discard

    // start hijacking
    h, _ := w.(http.Hijacker)
    lconn, _, _ := h.Hijack()
    defer lconn.Close()

    r.WriteProxy(rconn)
    Transfer(lconn, rconn)
    return
}

func (p *Proxyes)handleHTTP(w http.ResponseWriter, r *http.Request) {
    conn, err := net.Dial("tcp", p.proxy1) // Dial to upstream
    if err != nil {
	log.Println("net.Dial:", err)
	w.WriteHeader(http.StatusInternalServerError)
	return
    }
    defer conn.Close() // don't forget close

    if r.Header.Get("Proxy-Connection") == "Keep-Alive" {
	// close if Proxy-Connection exists
	r.Header.Set("Proxy-Connection", "close")
    }
    // always close
    r.Header.Set("Connection", "close")

    r.WriteProxy(conn)
    h, _ := w.(http.Hijacker)
    lconn, _, _ := h.Hijack()
    defer lconn.Close()

    io.Copy(lconn, conn)
}

func (p *Proxyes)Handler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodConnect {
	p.handleConnect(w,r)
    } else {
	p.handleHTTP(w, r)
    }
}

func main() {
    if len(os.Args) < 4 {
	log.Fatal("doubleproxy listen proxy1 proxy2")
    }
    p := &Proxyes{
	proxy1: os.Args[2],
	proxy2: os.Args[3],
    }
    listen := os.Args[1]
    http.ListenAndServe(listen, http.HandlerFunc(p.Handler))
}
