package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	logger := createLogger(proc)

	cs := []string {"cmd", "/c", procs[proc].cmdline}
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

func terminateProc(proc string) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	f, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	pid := procs[proc].cmd.Process.Pid
	f.Call(syscall.CTRL_C_EVENT, uintptr(pid))
	return nil
}
