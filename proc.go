package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"
)

// spawnProc starts the specified proc, and returns any error from running it.
func spawnProc(name string, errCh chan<- error) {
	proc := findProc(name)
	logger := createLogger(name, proc.colorIndex)

	cs := append(cmdStart, proc.cmdline)
	cmd := exec.Command(cs[0], cs[1:]...)
	fmt.Println("cmd", cmd.String())
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = procAttrs

	if proc.setPort {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", proc.port))
		fmt.Fprintf(logger, "Starting %s on port %d\n", name, proc.port)
	}
	if err := cmd.Start(); err != nil {
		select {
		case errCh <- err:
		default:
		}
		fmt.Fprintf(logger, "Failed to start %s: %s\n", name, err)
		return
	}
	proc.cmd = cmd
	proc.stoppedBySupervisor = false
	proc.mu.Unlock()
	err := cmd.Wait()
	proc.mu.Lock()
	proc.cond.Broadcast()
	if err != nil && !proc.stoppedBySupervisor {
		select {
		case errCh <- err:
		default:
		}
	}
	proc.waitErr = err
	proc.cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", name)
}

// stopProc is stopping the specified process. Issuing os.Kill if it does not terminate within 10 seconds. If signal is
// nil, os.Interrupt is used.
func stopProc(name string, signal os.Signal) error {
	if signal == nil {
		signal = os.Interrupt
	}
	proc := findProc(name)
	if proc == nil {
		return errors.New("unknown proc: " + name)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.cmd == nil {
		return nil
	}
	proc.stoppedBySupervisor = true

	err := terminateProc(proc, signal)
	if err != nil {
		return err
	}

	timeout := time.AfterFunc(10*time.Second, func() {
		proc.mu.Lock()
		defer proc.mu.Unlock()
		if proc.cmd != nil {
			err = killProc(proc.cmd.Process)
		}
	})
	proc.cond.Wait()
	timeout.Stop()
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(name string, wg *sync.WaitGroup, errCh chan<- error) error {
	proc := findProc(name)
	if proc == nil {
		return errors.New("unknown name: " + name)
	}

	proc.mu.Lock()
	if proc.cmd != nil {
		proc.mu.Unlock()
		return nil
	}

	if wg != nil {
		wg.Add(1)
	}
	go func() {
		spawnProc(name, errCh)
		if wg != nil {
			wg.Done()
		}
		proc.mu.Unlock()
	}()
	return nil
}

// stopProcs attempts to stop every running process and returns any non-nil
// error, if one exists. stopProcs will wait until all procs have had an
// opportunity to stop.
func stopProcs(sig os.Signal) error {
	var err error
	for _, proc := range procs {
		stopErr := stopProc(proc.name, sig)
		if stopErr != nil {
			err = stopErr
		}
	}
	return err
}

// spawn all procs.
func startProcs(sc <-chan os.Signal, exitOnError bool) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for _, proc := range procs {
		startProc(proc.name, &wg, errCh)
	}

	allProcsDone := make(chan struct{}, 1)
	if *exitOnStop {
		go func() {
			wg.Wait()
			allProcsDone <- struct{}{}
		}()
	}
	for {
		select {
		case err := <-errCh:
			if exitOnError {
				stopProcs(os.Interrupt)
				return err
			}
		case <-allProcsDone:
			return stopProcs(os.Interrupt)
		case sig := <-sc:
			return stopProcs(sig)
		}
	}
}

const sigint = unix.SIGINT
const sigterm = unix.SIGTERM
const sighup = unix.SIGHUP

var cmdStart = []string{"/bin/sh", "-c"}
var procAttrs = &unix.SysProcAttr{Setpgid: true}

func terminateProc(proc *procInfo, signal os.Signal) error {
	p := proc.cmd.Process
	if p == nil {
		return nil
	}

	pgid, err := unix.Getpgid(p.Pid)
	if err != nil {
		return err
	}

	// use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	pid := p.Pid
	if pgid == p.Pid {
		pid = -1 * pid
	}

	target, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return target.Signal(signal)
}

// killProc kills the proc with pid pid, as well as its children.
func killProc(process *os.Process) error {
	return unix.Kill(-1*process.Pid, unix.SIGKILL)
}

func notifyCh() <-chan os.Signal {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, sigterm, sigint, sighup)
	return sc
}
