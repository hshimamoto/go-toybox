// sshfwd
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./sshfwd <fwd config files...>
//
// config file format
// host destIP:port user password
// fwd localIP:port remoteIP:port

package main

import (
    "strings"
    "time"
    "log"
    "os"
    "io"
    "bufio"

    "net"
    "golang.org/x/crypto/ssh"
)

type Fwd struct {
    local, remote string
}

type Host struct {
    dest string
    user, pw string
    fwds []Fwd
}

func loadConfig(path string) []*Host {
    hosts := []*Host{}

    f, err := os.Open(path)
    if err != nil {
	return hosts
    }
    defer f.Close()

    var h *Host = nil
    sc := bufio.NewScanner(f)
    for sc.Scan() {
	line := strings.Trim(sc.Text(), "\n")
	a := strings.Split(line, " ")
	switch a[0] {
	case "host":
	    if len(a) < 4 {
		log.Println("bad line:", line)
		continue
	    }
	    h = &Host { dest: a[1], user: a[2], pw: a[3] }
	    hosts = append(hosts, h)
	case "fwd":
	    if len(a) < 3 {
		log.Println("bad line:", line)
		continue
	    }
	    h.fwds = append(h.fwds, Fwd { local: a[1], remote: a[2] })
	}
    }

    return hosts
}

func (fwd *Fwd)session(cli *ssh.Client, lconn *net.TCPConn) {
    defer lconn.Close()

    log.Println("forwarding", fwd.local, "to", fwd.remote)

    rconn, err := cli.Dial("tcp", fwd.remote)
    if err != nil {
	log.Println("ssh.Dial", err)
	return
    }
    defer rconn.Close()

    done1 := make(chan bool)
    done2 := make(chan bool)
    go func() {
	io.Copy(lconn, rconn)
	done1 <- true
    }()
    go func() {
	io.Copy(rconn, lconn)
	done2 <- true
    }()
    select {
    case <-done1: go func() { <-done2 }()
    case <-done2: go func() { <-done1 }()
    }

    time.Sleep(time.Second)

    log.Println("forwarding", fwd.local, "to", fwd.remote, "done")
}

func (h *Host)sshthread(cli *ssh.Client) {
    // create fwd listeners
    for _, fwd := range(h.fwds) {
	addr, err := net.ResolveTCPAddr("tcp", fwd.local)
	if err != nil {
	    log.Println("net.ResolveTCPAddr", err)
	    continue // TODO
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
	    log.Println("net.ListenTCP", err)
	    continue // TODO
	}
	log.Println("start listening on", fwd.local)
	go func() {
	    defer l.Close()
	    for {
		conn, err := l.AcceptTCP()
		if err != nil {
		    log.Println("AcceptTCP", err)
		    continue
		}
		go fwd.session(cli, conn)
	    }
	}()
    }
    // keepalive
    for {
	cli.Conn.SendRequest("keepalive@golang.org", true, nil)
	time.Sleep(time.Minute)
    }
}

func (h *Host)sshconnect() {
    cfg := &ssh.ClientConfig {
	User: h.user,
	Auth: []ssh.AuthMethod { ssh.Password(h.pw) },
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    cli, err := ssh.Dial("tcp", h.dest, cfg)
    if err != nil {
	log.Println("ssh.Dial", err)
	return
    }
    log.Println("ssh connection with", h.dest)
    go h.sshthread(cli)
}

func main() {
    if len(os.Args) == 1 {
	log.Println("./sshfwd <fwd config files...>")
	return
    }
    hosts := []*Host{}
    for _, path := range(os.Args) {
	hosts = append(hosts, loadConfig(path)...)
    }
    if len(hosts) == 0 {
	log.Println("No hosts")
	return
    }
    for _, h := range(hosts) {
	log.Println(h.dest, "w/", len(h.fwds), "fwds")
	h.sshconnect()
    }
    time.Sleep(time.Second)
    log.Println("service started. to stop the service, CTRL-C")
    for {
	time.Sleep(time.Hour) // don't care about sleep time, we will kill it anyway
    }
}
