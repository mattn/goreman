package main

import (
	"log"
	"strings"
	"sync"
)

const (
	Black = "\x1b[30m"
	Red = "\x1b[31m"
	Green = "\x1b[32m"
	Yellow = "\x1b[33m"
	Blue = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan = "\x1b[36m"
	White = "\x1b[37m"
)

type clogger struct {
	p string
}

var mutex = new(sync.Mutex)

func (l *clogger) Write(p []byte) (n int, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0:len(line)-2]
		}
		if line != "" {
			log.Printf("[%s] %s", l.p, line)
		}
	}
	n = len(p)
	return
}

func create_logger(proc string) *clogger {
	return &clogger {proc}
}
