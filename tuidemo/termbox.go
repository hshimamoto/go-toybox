// tuidemo / termbox.go
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// go run termbox.go

package main

import (
    "log"

    "github.com/nsf/termbox-go"
)

const cdef = termbox.ColorDefault

func eventloop(quit chan struct{}) {
    for {
	ev := termbox.PollEvent()
	switch ev.Type {
	case termbox.EventKey:
	    if ev.Key == termbox.KeyEnter {
		close(quit)
		return
	    }
	}
    }
}

func putmsg(x, y int, msg string) {
    rs := []rune(msg)
    for i, r := range rs {
	termbox.SetCell(x+i, y, r, cdef, cdef)
    }
}

func whitemsg(x, y int, msg string) {
    rs := []rune(msg)
    for i, r := range rs {
	termbox.SetCell(x+i, y, r, termbox.ColorWhite, cdef)
    }
}

func colorfulmsg(x, y int, msg string) {
    rs := []rune(msg)
    cols := []termbox.Attribute{
	termbox.ColorRed,
	termbox.ColorGreen,
	termbox.ColorYellow,
	termbox.ColorBlue,
	termbox.ColorMagenta,
	termbox.ColorCyan,
    }
    clen := len(cols)
    cidx := 0
    for i, r := range rs {
	termbox.SetCell(x+i, y, r, cols[cidx], cdef)
	cidx = (cidx + 1) % clen
    }
}

func main() {
    err := termbox.Init()
    if err != nil {
	log.Fatal(err)
    }
    defer termbox.Close()
    termbox.SetInputMode(termbox.InputEsc)

    // quit event
    quit := make(chan struct{})
    go eventloop(quit)

    termbox.Clear(cdef, cdef)
    whitemsg(0, 0, "Hello World")
    colorfulmsg(0, 1, "Colorful Message!")
    putmsg(0, 2, "Press Enter to Quit")
    termbox.SetCursor(0, 3)
    termbox.Flush()
    <-quit
}
