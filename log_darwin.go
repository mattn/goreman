package main

import (
	"github.com/mewkiz/pkg/term"
	"log"
	"strings"
	"sync"
)

type clogger struct {
	idx int
	proc string
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

// write handler of logger.
func (l *clogger) Write(p []byte) (n int, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0:len(line)-2]
		}
		if line != "" {
			log.Printf("[%s] %s", term.Color(l.proc, colors[l.idx]), term.Color(line, colors[l.idx]))
		}
	}
	n = len(p)
	return
}

// create logger instance.
func create_logger(proc string) *clogger {
	l := &clogger {ci, proc}
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
