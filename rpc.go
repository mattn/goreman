package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
)

type Goreman int

// rpc: start
func (r *Goreman) Start(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return startProc(proc)
}

// rpc: stop
func (r *Goreman) Stop(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return stopProc(proc, false)
}

// rpc: restart
func (r *Goreman) Restart(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return restartProc(proc)
}

// rpc: list
func (r *Goreman) List(empty string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	for proc := range procs {
		*ret += proc + "\n"
	}
	return err
}

// rpc: dump
func (r *Goreman) Dump(empty string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	for proc := range procs {
		if procs[proc].quit {
			*ret += proc + "\n"
		} else {
			*ret += "#" + proc + "\n"
		}
	}
	return err
}

// command: run.
func run(cmd, proc string) error {
	client, err := rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return err
	}
	defer client.Close()
	var ret string
	switch cmd {
	case "start":
		return client.Call("Goreman.Start", proc, &ret)
	case "stop":
		return client.Call("Goreman.Stop", proc, &ret)
	case "restart":
		return client.Call("Goreman.Restart", proc, &ret)
	case "list":
		err := client.Call("Goreman.List", "", &ret)
		fmt.Print(ret)
		return err
	case "dump":
		err := client.Call("Goreman.Dump", "", &ret)
		fmt.Print(ret)
		return err
	}
	return errors.New("Unknown command")
}

// start rpc server.
func startServer() error {
	gm := new(Goreman)
	rpc.Register(gm)
	server, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		return err
	}
	for {
		client, err := server.Accept()
		if err != nil {
			continue
		}
		rpc.ServeConn(client)
	}
	return nil
}
