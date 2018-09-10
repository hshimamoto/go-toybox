// cnc
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim: set sw=4 sts=4:
//
// Capture NetCat
//
// cnc <host> <port> [file]
//  or
// cnc dump <file>

package main

import (
    "encoding/binary"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "sync"
    "time"
)

var m *sync.Mutex
var ncdat io.Writer

func concat(dst io.Writer, src io.Reader, to bool, done chan struct{}) {
    var wg sync.WaitGroup

    for {
	buf := make([]byte, 4096)
	n, _ := src.Read(buf)
	tm := time.Now()
	// log it in background
	wg.Add(1)
	go func() {
	    m.Lock()
	    if to {
		ncdat.Write([]byte{ 1 })
	    } else {
		ncdat.Write([]byte{ 0 })
	    }
	    tbytes, _ := tm.MarshalBinary()
	    ncdat.Write(tbytes)
	    l := uint32(n)
	    binary.Write(ncdat, binary.LittleEndian, &l)
	    ncdat.Write(buf[:n])
	    m.Unlock()
	    wg.Done()
	}()
	if n <= 0 {
	    break
	}
	w, err := dst.Write(buf[:n])
	if w != n || err != nil {
	    break
	}
    }
    wg.Wait()
    close(done)
}

func dump(path string) {
    f, err := os.Open(path)
    if err != nil {
	log.Fatal("os.Open:", err)
	return
    }
    defer f.Close()

    to := []byte{ 0 }
    tbytes := make([]byte, 15)

    for {
	n, _ := f.Read(to)
	if n <= 0 {
	    break
	}
	n, _ = f.Read(tbytes)
	if n < 15 {
	    break
	}
	tm := &time.Time{}
	tm.UnmarshalBinary(tbytes)
	var l uint32
	binary.Read(f, binary.LittleEndian, &l)
	buf := make([]byte, l)
	if l > 0 {
	    f.Read(buf)
	}
	arrow := "<- "
	if to[0] == 1 { arrow = " ->" }
	fmt.Printf("%v %s len=%d\n", tm, arrow, l)
	c := int(l)
	off := 0
	for c > 0 {
	    sz := 16
	    if sz > c {
		sz = c
	    }
	    line := fmt.Sprintf("% x", buf[:sz])
	    fmt.Printf(" %04x| %s\n", off, line)
	    buf = buf[sz:]
	    off += sz
	    c -= sz
	}
    }
}

func main() {
    if len(os.Args) < 3 {
	usage := "cnc <host> <port> [file] | cnc dump <file>"
	log.Fatal(usage)
	return
    }

    if os.Args[1] == "dump" {
	dump(os.Args[2])
	return
    }

    // init mutex
    m = &sync.Mutex{}

    hostport := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])

    conn, err := net.Dial("tcp", hostport)
    if err != nil {
	log.Fatal("net.Dial:", err)
	return
    }
    defer conn.Close()

    path := fmt.Sprintf("ncdat.%d", os.Getpid())
    if len(os.Args) > 3 {
	path = os.Args[3]
    }

    f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0644)
    if err != nil {
	log.Fatal("os.OpenFilea:", err)
	return
    }
    defer f.Close()

    ncdat = f

    d1 := make(chan struct{})
    d2 := make(chan struct{})
    go concat(os.Stdin, conn, false, d1)
    go concat(conn, os.Stdout, true, d2)
    <-d1
    <-d2
}
