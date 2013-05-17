package main

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

func done() {
	func() {
		defer func() {
			recover()
		}()
		wg.Done()
	}()
}

// stop specified proc.
func stopProc(proc string, quit bool) error {
	if _, ok := procs[proc]; !ok {
		return errors.New("Unknown proc: " + proc)
	}
	if procs[proc].cmd == nil {
		return nil
	}

	procs[proc].quit = quit
	err := procs[proc].cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}
	_, err = procs[proc].cmd.Process.Wait()
	if err == nil {
		procs[proc].cmd = nil
	}
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	if _, ok := procs[proc]; !ok {
		return errors.New("Unknown proc: " + proc)
	}
	if procs[proc].cmd != nil {
		return nil
	}

	go func() {
		if spawnProc(proc) {
			done()
		}
	}()
	return nil
}

// restart specified proc.
func restartProc(proc string) error {
	if _, ok := procs[proc]; !ok {
		return errors.New("Unknown proc: " + proc)
	}
	stopProc(proc, false)
	time.Sleep(1 * time.Second)
	return startProc(proc)
}

// spawn all procs.
func startProcs() error {
	wg.Add(len(procs))
	for proc := range procs {
		startProc(proc)
	}
	sc := make(chan os.Signal, 10)
	state := true
	go func() {
		wg.Wait()
		state = false
		sc <- syscall.SIGINT
	}()
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-sc
	for proc, p := range procs {
		if p.cmd != nil {
			stopProc(proc, true)
		}
		done()
	}
	return nil
}
