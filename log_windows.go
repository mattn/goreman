package main

import (
	"log"
	"os"
	"strings"
	"sync"
	"github.com/anschelsc/doscolor"
)

type clogger struct {
	i int
	p string
}

var mutex = new(sync.Mutex)
var cerr = doscolor.NewWrapper(os.Stderr)
var elog = log.New(cerr, "", log.LstdFlags)
var colors = []doscolor.Color {
	doscolor.Green | doscolor.Bright,
	doscolor.Cyan | doscolor.Bright,
	doscolor.Magenta | doscolor.Bright,
	doscolor.Yellow | doscolor.Bright,
	doscolor.Blue | doscolor.Bright,
	doscolor.Red | doscolor.Bright,
}
var ci int

func (l *clogger) Write(p []byte) (n int, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0:len(line)-2]
		}
		if line != "" {
			cerr.Set(colors[l.i])
			elog.Printf("[%s] %s", l.p, line)
			cerr.Restore()
		}
	}
	n = len(p)
	return
}

func create_logger(proc string) *clogger {
	l := &clogger {ci, proc}
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
