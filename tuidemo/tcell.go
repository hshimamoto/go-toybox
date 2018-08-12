// tuidemo / tcell.go
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// go run tcell.go

package main

import (
    "log"

    "github.com/gdamore/tcell"
)

func eventloop(s tcell.Screen, quit chan struct{}) {
    for {
	ev := s.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
	    if ev.Key() == tcell.KeyEnter {
		close(quit)
		return
	    }
	}
    }
}

func putmsg(s tcell.Screen, x, y int, msg string) {
    rs := []rune(msg)
    for i, r := range rs {
	s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
    }
}

func whitemsg(s tcell.Screen, x, y int, msg string) {
    rs := []rune(msg)
    for i, r := range rs {
	s.SetContent(x+i, y, r, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
    }
}

func colorfulmsg(s tcell.Screen, x, y int, msg string) {
    rs := []rune(msg)
    cols := []tcell.Style{
	tcell.StyleDefault.Foreground(tcell.ColorRed),
	tcell.StyleDefault.Foreground(tcell.ColorGreen),
	tcell.StyleDefault.Foreground(tcell.ColorYellow),
	tcell.StyleDefault.Foreground(tcell.ColorBlue),
	tcell.StyleDefault.Foreground(tcell.ColorPurple),
	tcell.StyleDefault.Foreground(tcell.ColorAqua),
    }
    clen := len(cols)
    cidx := 0
    for i, r := range rs {
	s.SetContent(x+i, y, r, nil, cols[cidx])
	cidx = (cidx + 1) % clen
    }
}

func main() {
    screen, err := tcell.NewScreen()
    if err != nil {
	log.Fatal(err)
    }
    err = screen.Init()
    if err != nil {
	log.Fatal(err)
    }
    defer screen.Fini()

    // quit event
    quit := make(chan struct{})
    go eventloop(screen, quit)

    //screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
    screen.Clear()
    whitemsg(screen, 0, 0, "Hello World")
    colorfulmsg(screen, 0, 1, "Colorful Message!")
    putmsg(screen, 0, 2, "Press Enter to Quit")
    screen.ShowCursor(0, 3)
    screen.Show()
    <-quit
}
