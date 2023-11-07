package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/sys/windows"
)

var cmdStart = []string{"cmd", "/c"}
var procAttrs = &windows.SysProcAttr{
	CreationFlags: windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_NEW_PROCESS_GROUP,
}

func terminateProc(proc *procInfo, _ os.Signal) error {
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	defer dll.Release()

	pid := proc.cmd.Process.Pid

	f, err := dll.FindProc("AttachConsole")
	if err != nil {
		return err
	}
	r1, _, err := f.Call(uintptr(pid))
	if r1 == 0 && err != syscall.ERROR_ACCESS_DENIED {
		return err
	}

	f, err = dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return err
	}
	r1, _, err = f.Call(0, 1)
	if r1 == 0 {
		return err
	}
	f, err = dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r1, _, err = f.Call(windows.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		return err
	}
	r1, _, err = f.Call(windows.CTRL_C_EVENT, uintptr(pid))
	if r1 == 0 {
		return err
	}
	return nil
}

func killProc(process *os.Process) error {
	return process.Kill()
}

func notifyCh() <-chan os.Signal {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, os.Interrupt)
	return sc
}

// This is a no-op on Windows.
func startPTY(logger *clogger, cmd *exec.Cmd) error {
	cmd.Stdout = logger
	cmd.Stderr = logger
}
