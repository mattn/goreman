package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

var procs map[string]*exec.Cmd

func start(proc []string) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	f, err := dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return err
	}
	r, _, err := f.Call(0, 1)
	if r == 0 {
		return err
	}

	entry, err := getEntry()
	if err != nil {
		return err
	}
	if len(proc) != 0 {
		tmp := map[string]string {}
		for _, v := range proc {
			tmp[v] = entry[v]
		}
		entry = tmp
	}

	var wg sync.WaitGroup
	wg.Add(len(entry))
	for k, v := range entry {
		go func(k string, v string) {
			log.Printf("[%s] START", k)
			cs := []string {"cmd", "/c", v}
			cmd := exec.Command(cs[0], cs...)
			cmd.Stdin = nil
			cmd.Stdout = &logger{k}
			cmd.Stderr = &logger{k}

			err = cmd.Start()
			if err != nil {
				log.Fatal("failed to execute external command. %s", err)
				os.Exit(1)
			}

			cmd.Wait()
			wg.Done()
			log.Printf("[%s] QUIT", k)
		}(k, v)
	}

	go func() {
		sc := make(chan os.Signal, 10)
		signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, os.Interrupt)
		for sig := range sc {
			println(sig)
		}
	}()

	wg.Wait()
	return nil
}
