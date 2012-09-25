// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	logger := createLogger(proc)

	cs := []string{"/bin/bash", "-c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger

	fmt.Fprintf(logger, "START")
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "failed to execute external command. %s", err)
		return true
	}
	procs[proc].cmd = cmd
	procs[proc].quit = true
	procs[proc].cmd.Wait()
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "QUIT")

	return procs[proc].quit
}

// stop specified proc.
func stopProc(proc string, quit bool) error {
	if procs[proc].cmd == nil {
		return nil
	}

	procs[proc].quit = quit
	pid := procs[proc].cmd.Process.Pid

	syscall.Kill(pid, syscall.SIGINT)
	return nil
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	if procs[proc].cmd != nil {
		return nil
	}

	go func() {
		if spawnProc(proc) {
			wg.Done()
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
	return startProc(proc)
}

// spawn all procs.
func startProcs() error {
	wg.Add(len(procs))
	for proc := range procs {
		startProc(proc)
	}
	sc := make(chan os.Signal, 10)
	done := false
	go func() {
		wg.Wait()
		done = true
		sc <- syscall.SIGINT
	}()
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-sc
	if !done {
		for proc, p := range procs {
			if p.cmd != nil {
				stopProc(proc, true)
			} else {
				wg.Done()
			}
		}
	}
	wg.Wait()
	return nil
}
