// go-toybox/fwd
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "fmt"
    "log"
    "net"
    "os"

    "github.com/hshimamoto/go-session"
    "github.com/hshimamoto/go-iorelay"
)

func show_usage() {
    fmt.Println("Version: fwd v1.0.0 in go-toyboxv")
    fmt.Println("Usage: fwd <listen> <dest>")
}

func main() {
    if len(os.Args) < 3 {
	show_usage()
	return
    }
    s, err := session.NewServer(os.Args[1], func(conn net.Conn) {
	defer conn.Close()
	fconn, err := session.Dial(os.Args[2])
	if err != nil {
	    log.Println("Dial %s %v\n", os.Args[2], err)
	    return
	}
	defer fconn.Close()
	iorelay.Relay(conn, fconn)
    })
    if err != nil {
	log.Printf("session.NewServer: %v\n", err)
	return
    }
    s.Run()
}
