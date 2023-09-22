//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

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

func startPTY(logger *clogger, cmd *exec.Cmd) error {
	if *usePty {
		p, t, err := pty.Open()
		if err != nil {
			return fmt.Errorf("failed to open PTY: %w", err)
		}
		defer p.Close()
		defer t.Close()
		cmd.Stdout = t
		cmd.Stderr = t
		go io.Copy(logger, p)
	} else {
		cmd.Stdout = logger
		cmd.Stderr = logger
	}
	return nil
}
