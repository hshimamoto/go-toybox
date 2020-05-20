// go-toybox/fwd
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "strconv"

    "github.com/hshimamoto/go-session"
    "github.com/hshimamoto/go-iorelay"
    "github.com/mxk/go-flowrate/flowrate"
)

func show_usage() {
    fmt.Println("Version: fwd v2.0.0 in go-toyboxv")
    fmt.Println(
`Usage: fwd <listen> <dest> [KiB/s or MiB/s]
ex)
	fwd :8080 www.example.com:80
	fwd :8080 www.example.com:80 100K
	fwd :8080 www.example.com:80 10M`)
}

type FlowrateIO struct {
    r *flowrate.Reader
    w *flowrate.Writer
}

func (f *FlowrateIO)Read(p []byte) (int, error) {
    return f.r.Read(p)
}

func (f *FlowrateIO)Write(p []byte) (int, error) {
    return f.w.Write(p)
}

func main() {
    if len(os.Args) < 3 {
	show_usage()
	return
    }
    var mb int64 = 0
    if len(os.Args) > 3 {
	lim := os.Args[3]
	unit := 0
	switch lim[len(lim) - 1] {
	case 'k': unit = 1024
	case 'K': unit = 1024
	case 'm': unit = 1024 * 1024
	case 'M': unit = 1024 * 1024
	}
	m, err := strconv.Atoi(lim[:len(lim) - 1])
	if err != nil {
	    log.Printf("Atoi %v", err)
	    m = 0
	}
	mb = int64(m * unit)
	if mb == 0 {
	    log.Printf("bad option: %s\n", lim)
	    return
	}
	log.Printf("Set Ratelimit %d bytes/s", mb)
    }
    s, err := session.NewServer(os.Args[1], func(conn net.Conn) {
	defer conn.Close()
	fconn, err := session.Dial(os.Args[2])
	if err != nil {
	    log.Println("Dial %s %v\n", os.Args[2], err)
	    return
	}
	defer fconn.Close()
	if mb > 0 {
	    f1 := &FlowrateIO{
		r: flowrate.NewReader(conn, mb),
		w: flowrate.NewWriter(conn, mb),
	    }
	    f2 := &FlowrateIO{
		r: flowrate.NewReader(fconn, mb),
		w: flowrate.NewWriter(fconn, mb),
	    }
	    iorelay.Relay(f1, f2)
	} else {
	    iorelay.Relay(conn, fconn)
	}
    })
    if err != nil {
	log.Printf("session.NewServer: %v\n", err)
	return
    }
    s.Run()
}
