// go-toybox/bashhistcompact
// MIT License Copyright(c) 2020 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "time"
)

func main() {
    // try to create backup file ~/.bashhist
    home, err := os.UserHomeDir()
    if err != nil {
	log.Fatal(err)
	return
    }
    pid := os.Getpid()
    // format YYYYMMDDhhmmss
    now := time.Now().Format("20060102150405")
    bkname := fmt.Sprintf("%s/.bashhist/histroy-%s-%d", home, now, pid)
    bk, err := os.OpenFile(bkname, os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
	log.Fatal(err)
	return
    }
    defer bk.Close()

    histname := fmt.Sprintf("%s/.bash_history", home)

    // read everything
    hist, err := ioutil.ReadFile(histname)
    if err != nil {
	log.Fatal(err)
	return
    }

    // write everything as backup
    bk.Write(hist)

    // craft history
    f, err := os.OpenFile(histname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
    if err != nil {
	log.Fatal(err)
	return
    }
    defer f.Close()

    prev := ""
    for _, line := range strings.Split(string(hist), "\n") {
	// remove simple command
	if line == "ls" {
	    continue
	}
	if line == "pwd" {
	    continue
	}
	if line == "cd" {
	    continue
	}
	if line == "cd .." {
	    continue
	}
	// remove job control
	if line == "jobs" {
	    continue
	}
	if line == "%" {
	    continue
	}
	// remove catastrophic
	if line == "rm -rf *" {
	    continue
	}
	// remove same line
	if line == prev {
	    continue
	}
	if line == "" {
	    continue
	}
	f.Write([]byte(line + "\n"))
	prev = line
    }
}
