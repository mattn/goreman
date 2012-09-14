package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

func spawn_proc(proc string) bool {
	logger := create_logger(proc)

	cs := []string {"cmd", "/c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger

	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	fmt.Fprintf(logger, "[%s] START", proc)
	err := cmd.Start()
	if err != nil {
		log.Fatal("failed to execute external command. %s", err)
		return true
	}
	procs[proc].cmd = cmd
	procs[proc].quit = true
	procs[proc].cmd.Wait()
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "[%s] QUIT", proc)

	return procs[proc].quit
}

func stop_proc(proc string, quit bool) error {
	if procs[proc].cmd == nil {
		return nil
	}

	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	f, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}

	procs[proc].quit = quit
	pid := procs[proc].cmd.Process.Pid
	f.Call(syscall.CTRL_C_EVENT, uintptr(pid))
	return nil
}

func start_proc(proc string) error {
	if procs[proc].cmd != nil {
		return nil
	}

	go func() {
		if spawn_proc(proc) {
			wg.Done()
		}
	}()
	return nil
}

func restart_proc(proc string) error {
	err := stop_proc(proc, false)
	if err != nil {
		return err
	}
	return start_proc(proc)
}

func start_procs() error {
	wg.Add(len(procs))
	for proc := range procs {
		start_proc(proc)
	}
	go func() {
		sc := make(chan os.Signal, 10)
		signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		for _ = range sc {
			for k, v := range procs {
				if v == nil {
					wg.Done()
				} else {
					stop_proc(k, true)
				}
			}
			break
		}
	}()
	wg.Wait()
	return nil
}
