// +build !windows

package main

import (
	"fmt"
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

func terminateProc(proc string) error {
	return procs[proc].cmd.Process.Signal(syscall.SIGHUP)
}
