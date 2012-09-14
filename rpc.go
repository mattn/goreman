package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
)

type Goreman int

func (r *Goreman) Start(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return start_proc(proc)
}

func (r *Goreman) Stop(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return stop_proc(proc, false)
}

func (r *Goreman) Restart(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return restart_proc(proc)
}

func run(cmd, proc string) error {
	client, err := rpc.Dial("tcp", "127.0.0.1:5555")
	if err != nil {
		return err
	}
	var ret string
	switch cmd {
	case "start":
		return client.Call("Goreman.Start", proc, &ret)
	case "stop":
		return client.Call("Goreman.Stop", proc, &ret)
	case "restart":
		return client.Call("Goreman.Restart", proc, &ret)
	}
	return errors.New("Unknown command")
}

func start_server() error {
	gm := new(Goreman)
	rpc.Register(gm)
	server, err := net.Listen("tcp", "0.0.0.0:5555")
	if err != nil {
		return err
	}
	for {
		client, err := server.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		rpc.ServeConn(client)
	}
	return nil
}
