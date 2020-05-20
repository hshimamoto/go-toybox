// go-toybox/fwdset
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"
    "github.com/BurntSushi/toml"
)

// Example config.toml
//
// [[Fwds]]
// Name = "name"
// Src = ":8080"
// Dst = "proxy:8080"
// Limit = "10M"

type fwd struct {
    Name, Src, Dst, Limit string
}

type fwdconfig struct {
    Fwds []fwd
}

func loadconfig(config string) (*fwdconfig, error) {
    cfg := &fwdconfig{}
    if _, err := toml.DecodeFile(config, cfg); err != nil {
	log.Printf("config %s error: %v\n", config, err)
	return nil, err
    }
    return cfg, nil
}

type fwdproc struct {
    fwd
    cmd *exec.Cmd
}

func (fp *fwdproc)run(w chan *fwdproc) {
    if fp.Limit == "" {
	fp.cmd = exec.Command("fwd", fp.Src, fp.Dst)
    } else {
	fp.cmd = exec.Command("fwd", fp.Src, fp.Dst, fp.Limit)
    }
    go func() {
	fp.cmd.Run()
	time.Sleep(time.Second)
	w <- fp // send signal
	// clear cmd
	fp.cmd = nil
    }()
}

func manage(config string) {
    cfg, err := loadconfig(config)
    if err != nil {
	return
    }
    // create fwds
    fwds := []*fwdproc{}
    w := make(chan *fwdproc)
    for _, fwd := range(cfg.Fwds) {
	f := &fwdproc{
	    fwd: fwd,
	    cmd: nil,
	}
	fwds = append(fwds, f)
    }
    for {
	// check state
	for _, fp := range(fwds) {
	    if fp.cmd == nil {
		fmt.Printf("start %s %s %s %s\n", fp.Name, fp.Src, fp.Dst, fp.Limit)
		// process restart
		fp.run(w)
	    }
	}
	select {
	case fp := <-w:
	    fmt.Printf("done %s %s %s %s\n", fp.Name, fp.Src, fp.Dst, fp.Limit)
	case <-time.After(time.Minute): // just for timeout
	}
	// reload config
	cfg, err := loadconfig(config)
	if err != nil {
	    continue
	}
	// check renewal?
	newfwds := []*fwdproc{}
	for _, fwd := range(cfg.Fwds) {
	    // find the same
	    found := false
	    for _, fp := range(fwds) {
		if (fp.Name == fwd.Name) &&
			(fp.Src == fwd.Src) && (fp.Dst == fwd.Dst) && (fp.Limit == fwd.Limit) {
		    // keep it
		    newfwds = append(newfwds, fp)
		    found = true
		    break
		}
	    }
	    if found {
		continue
	    }
	    f := &fwdproc{
		fwd: fwd,
		cmd: nil,
	    }
	    newfwds = append(newfwds, f)
	}
	fwds = newfwds
    }
}

func main() {
    if len(os.Args) != 2 {
	log.Println("fwdset <config toml>")
	return
    }
    manage(os.Args[1])
}
