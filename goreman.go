package main

import (
	"fmt"
	"errors"
	"os/exec"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
)

const version = "0.0.1"

const (
	Black = "\x1b[30m"
	Red = "\x1b[31m"
	Green = "\x1b[32m"
	Yellow = "\x1b[33m"
	Blue = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan = "\x1b[36m"
	White = "\x1b[37m"
)

func usage() {
	println(`
Tasks:
  goreman check                  # Validate your application's Procfile
  goreman export FORMAT LOCATION # Export the application to another process...

  goreman help [TASK]            # Describe available tasks or one specific ...

  goreman run COMMAND [ARGS...]  # Run a command using your application's en...

  goreman start [PROCESS]        # Start the application (or a specific PROC...

  goreman version                # Display Goreman version

Options:
  -f, [--procfile=PROCFILE]  # Default: Procfile
  -d, [--root=ROOT]          # Default: Procfile directory
`[1:])
	os.Exit(0)
}

type logger struct {
	p string
}

func (l *logger) Write(p []byte) (n int, err error) {
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[0:len(line)-2]
		}
		if line != "" {
			log.Printf("[%s] %s", l.p, line)
		}
	}
	n = len(p)
	return
}

type Goreman int

func (r *Goreman) Start(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return start(proc)
}

func (r *Goreman) Stop(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return stop(proc, false)
}

func (r *Goreman) Restart(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return restart(proc)
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

type proc_info struct {
	p string
	l string
	q bool
	c *exec.Cmd
}
var procs map[string]*proc_info
var entry map[string]string

func main() {
	if len(os.Args) == 1 {
		usage()
	}
	cmd := os.Args[1]

	var content []byte
	var err error
	switch cmd {
	case "check":
		println("not implemented")
		break
	case "export":
		println("not implemented")
		break
	case "help":
		usage()
		break
	case "run":
		if len(os.Args) != 4 {
			usage()
		}
		err = run(os.Args[2], os.Args[3])
		break
	case "start":
		entry = map[string]string {}
		procs = map[string]*proc_info {}
		content, err = ioutil.ReadFile("Procfile")
		if err != nil {
			break
		}
		for _, line := range strings.Split(string(content), "\n") {
			tokens := strings.SplitN(line, ":", 2)
			if len(tokens) == 2 && tokens[0][0] != '#' {
				entry[strings.TrimSpace(tokens[0])] = strings.TrimSpace(tokens[1])
			}
		}
		if len(entry) == 0 {
			err = errors.New("No valid entry")
			break
		}
		go func() {
			gm := new(Goreman)
			rpc.Register(gm)
			server, err := net.Listen("tcp", "0.0.0.0:5555")
			if err != nil {
				return
			}
			for {
				client, err := server.Accept()
				if err != nil {
					log.Println(err.Error())
					continue
				}
				rpc.ServeConn(client)
			}
		}()
		err = start_procs(os.Args[2:])
		break
	case "version":
		fmt.Println(version)
		break
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
