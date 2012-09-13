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

func create_proc(proc string, cmdline string, logger *clogger) *proc_info {
	cs := []string {"cmd", "/c", cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger

	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	err := cmd.Start()
	if err != nil {
		log.Fatal("failed to execute external command. %s", err)
		return nil
	}
	return &proc_info { proc, cmdline, true, cmd, logger }
}

func stop(proc string, quit bool) error {
	if procs[proc] == nil {
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

	procs[proc].q = quit
	pid := procs[proc].c.Process.Pid
	f.Call(syscall.CTRL_C_EVENT, uintptr(pid))
	return nil
}

func start(proc string) error {
	if procs[proc] != nil {
		return nil
	}

	go func(k string, v string) {
		l := create_logger(k)
		fmt.Fprintf(l, "[%s] START", k)
		procs[k] = create_proc(k, v, l)
		procs[k].c.Wait()
		q := procs[k].q
		procs[k] = nil
		fmt.Fprintf(l, "[%s] QUIT", k)
		if q {
			wg.Done()
		}
	}(proc, entry[proc])
	return nil
}

func restart(proc string) error {
	err := stop(proc, false)
	if err != nil {
		return err
	}
	return start(proc)
}

func start_procs(proc []string) error {
	cerr.Save()

	procs = map[string]*proc_info {}
	if len(proc) != 0 {
		tmp := map[string]string {}
		for _, v := range proc {
			tmp[v] = entry[v]
		}
		entry = tmp
	}

	wg.Add(len(entry))
	for k := range entry {
		start(k)
	}

	go func() {
		sc := make(chan os.Signal, 10)
		signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		for _ = range sc {
			for k, v := range procs {
				if v == nil {
					wg.Done()
				} else {
					stop(k, true)
				}
			}
			break
		}
	}()

	wg.Wait()
	return nil
}
