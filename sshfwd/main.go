// sshfwd
// MIT License Copyright(c) 2018, 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./sshfwd <fwd config files...>
//
// config file format
// host destIP:port user keyfile or password
// fwd localIP:port remoteIP:port

package main

import (
    "io/ioutil"
    "strings"
    "time"
    "log"
    "os"
    "bufio"

    "net"
    "golang.org/x/crypto/ssh"
    "github.com/hshimamoto/go-iorelay"
    "github.com/hshimamoto/go-session"
)

type Fwd struct {
    local, remote string
}

type Host struct {
    proxy string
    dest string
    user, key string
    fwds []Fwd
}

func loadConfig(path string) []*Host {
    hosts := []*Host{}
    proxy := ""

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
	case "proxy":
	    proxy = a[1]
	case "host":
	    if len(a) < 4 {
		log.Println("bad line:", line)
		continue
	    }
	    h = &Host { proxy: proxy, dest: a[1], user: a[2], key: a[3] }
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

func (fwd *Fwd)session(cli *ssh.Client, lconn net.Conn) {
    defer lconn.Close()

    log.Println("forwarding", fwd.local, "to", fwd.remote)

    rconn, err := cli.Dial("tcp", fwd.remote)
    if err != nil {
	log.Println("ssh.Dial", err)
	return
    }
    defer rconn.Close()

    iorelay.Relay(lconn, rconn)

    time.Sleep(time.Second)

    log.Println("forwarding", fwd.local, "to", fwd.remote, "done")
}

func (h *Host)sshthread(cli *ssh.Client) {
    // create fwd listeners
    for _, f := range(h.fwds) {
	fwd := f
	s, err := session.NewServer(fwd.local, func(conn net.Conn) {
	    fwd.session(cli, conn)
	})
	if err != nil {
	    log.Println("session.NewServer", err)
	    continue // TODO
	}
	log.Println("start listening on", fwd.local)
	go s.Run()
    }
    // keepalive
    for {
	cli.Conn.SendRequest("keepalive@golang.org", true, nil)
	time.Sleep(time.Minute)
    }
}

func (h *Host)sshconnect() {
    cfg := &ssh.ClientConfig{
	User: h.user,
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    buf, err := ioutil.ReadFile(h.key)
    if err != nil {
	// no keyfile use password
	cfg.Auth = []ssh.AuthMethod{ ssh.Password(h.key) }
    } else {
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
	    log.Printf("ParsePrivateKey: %s %v\n", h.key, err)
	    return
	}
	cfg.Auth = []ssh.AuthMethod{ ssh.PublicKeys(key) }
    }

    var conn net.Conn
    if h.proxy == "" {
	conn, err = session.Dial(h.dest)
	if err != nil {
	    log.Printf("Dial %s: %v\n", h.dest, err)
	    return
	}
    } else {
	conn, err = session.Dial(h.proxy)
	if err != nil {
	    log.Printf("Dial proxy %s: %v\n", h.proxy, err)
	    return
	}
	conn.Write([]byte("CONNECT " + h.dest + " HTTP/1.1\r\n\r\n"))
	buf := make([]byte, 256)
	conn.Read(buf) // discard HTTP/1.1 200 Established
    }
    // start ssh through conn
    cconn, cchans, creqs, err := ssh.NewClientConn(conn, h.dest, cfg)
    if err != nil {
	log.Printf("NewClientConn %s: %v\n", h.dest, err)
	conn.Close()
	return
    }
    cli := ssh.NewClient(cconn, cchans, creqs)
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
