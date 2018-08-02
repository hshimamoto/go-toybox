// simple-proxy
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./httpcat URL
//

package main

import (
    "net/http"
    "os"
)

func main() {
    if len(os.Args) < 2 {
	return
    }
    url := os.Args[1]
    resp, err := http.Get(url)
    if err != nil {
	return
    }
    body := resp.Body
    defer body.Close()
    for {
	buf := make([]byte, 8192)
	n, err := body.Read(buf)
	if n > 0 {
	    os.Stdout.Write(buf[:n])
	}
	if err != nil {
	    break
	}
    }
}
