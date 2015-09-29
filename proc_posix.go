// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	logger := createLogger(proc)

	cs := []string{"/bin/bash", "-c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procs[proc].port))

	fmt.Fprintf(logger, "Starting %s on port %d\n", proc, procs[proc].port)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return true
	}
	procs[proc].cmd = cmd
	procs[proc].quit = true
	procs[proc].cmd.Wait()
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)

	return procs[proc].quit
}

func terminateProc(proc string) error {
	p := procs[proc].cmd.Process
	if p == nil {
		return nil
	}
	// find pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	group, err := os.FindProcess(-1 * p.Pid)
	if err == nil {
		err = group.Signal(syscall.SIGHUP)
	}
	return err
}
