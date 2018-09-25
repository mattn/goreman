package main

import (
	"errors"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Stop the specified proc, issuing SIGKILL if it does not terminate within 10
// seconds. If signal is nil, SIGTERM is used.
func stopProc(proc string, signal os.Signal) error {
	if signal == nil {
		signal = os.Interrupt
	}
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil {
		return nil
	}

	err := terminateProc(proc, signal)
	if err != nil {
		return err
	}

	timeout := time.AfterFunc(10*time.Second, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p, ok := procs[proc]; ok && p.cmd != nil {
			err = killProc(p.cmd.Process)
		}
	})
	p.cond.Wait()
	timeout.Stop()
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	p.mu.Lock()
	if procs[proc].cmd != nil {
		p.mu.Unlock()
		return nil
	}

	wg.Add(1)
	go func() {
		spawnProc(proc)
		wg.Done()
		p.mu.Unlock()
	}()
	return nil
}

// restart specified proc.
func restartProc(proc string) error {
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	stopProc(proc, nil)
	return startProc(proc)
}

// stopProcs attempts to stop every running process and returns any non-nil
// error, if one exists. stopProcs will wait until all procs have had an
// opportunity to stop.
func stopProcs(sig os.Signal) error {
	var err error
	for proc := range procs {
		stopErr := stopProc(proc, sig)
		if stopErr != nil {
			err = stopErr
		}
	}
	return err
}

// spawn all procs.
func startProcs() error {
	for proc := range procs {
		startProc(proc)
	}
	allProcsDone := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		allProcsDone <- struct{}{}
	}()
	sc := notifyCh()
	for {
		select {
		// TODO: add more events here.
		case <-allProcsDone:
			return stopProcs(os.Interrupt)
		case sig := <-sc:
			return stopProcs(sig)
		}
	}
	return nil
}
