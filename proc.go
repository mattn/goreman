package main

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

// stop specified proc.
func stopProc(proc string, quit bool) error {
	if procs[proc].cmd == nil {
		return nil
	}

	procs[proc].quit = quit
	err := procs[proc].cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}
	_, err = procs[proc].cmd.Process.Wait()
	return err
}

func done() {
	func() {
		defer func() {
			recover()
		}()
		wg.Done()
	}()
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
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
	err := stopProc(proc, false)
	if err != nil {
		return err
	}
	println("spawn")
	if !spawnProc(proc) {
		return errors.New("Failed to restart")
	}
	println("spawned")
	return nil
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
