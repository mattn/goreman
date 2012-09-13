package main

import (
	"log"
	"strings"
	"sync"
	"github.com/mewkiz/pkg/term"
)

type clogger struct {
	i int
	p string
}
var colors = []string {
	term.FgGreen,
	term.FgCyan,
	term.FgMagenta,
	term.FgYellow,
	term.FgBlue,
	term.FgRed,
}
var ci int

var mutex = new(sync.Mutex)

func (l *clogger) Write(p []byte) (n int, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0:len(line)-2]
		}
		if line != "" {
			term.Color(os.Stderr, colors[l.i])
			log.Printf("[%s] %s", l.p, line)
			term.Color(os.Stderr, term.FgWhite)
		}
	}
	n = len(p)
	return
}

func create_logger(proc string) *clogger {
	l := &clogger {proc}
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
