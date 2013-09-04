package main

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"strings"
	"sync"
	"time"
)

type clogger struct {
	idx  int
	proc string
}

var colors = []ct.Color{
	ct.Green,
	ct.Cyan,
	ct.Magenta,
	ct.Yellow,
	ct.Blue,
	ct.Red,
}
var ci int

var mutex = new(sync.Mutex)

// write handler of logger.
func (l *clogger) Write(p []byte) (n int, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0 : len(line)-2]
		}
		if line != "" {
			ct.ChangeColor(colors[l.idx], false, ct.None, false)
			now := time.Now().Format("15:04:05")
			format := fmt.Sprintf("%%s %%%ds | ", maxProcNameLength)
			fmt.Printf(format, now, l.proc)
			ct.ResetColor()
			fmt.Println(line)
		}
	}
	n = len(p)
	return
}

// create logger instance.
func createLogger(proc string) *clogger {
	l := &clogger{ci, proc}
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
