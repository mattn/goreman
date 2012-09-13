package main

import (
	"flag"
	"fmt"
	"errors"
	"os/exec"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"sort"
	"strings"
)

const version = "0.0.1"

func usage() {
	println(`
Tasks:
  goreman check                  # Show entries in Procfile
  goreman help [TASK]            # Show this help
  goreman run COMMAND [ARGS...]  # Run a command (start/stop/restart)
  goreman start [PROCESS]        # Start the application
  goreman version                # Display Goreman version

Options:
  -f # Default: Procfile
  -d # Default: Procfile directory
`[1:])
	os.Exit(0)
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

var entry map[string]string

type proc_info struct {
	p string
	l string
	q bool
	c *exec.Cmd
	w *clogger
}
var procs map[string]*proc_info

func readEntry() error {
	entry = map[string]string {}
	content, err := ioutil.ReadFile(*procfile)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) == 2 && tokens[0][0] != '#' {
			entry[strings.TrimSpace(tokens[0])] = strings.TrimSpace(tokens[1])
		}
	}
	if len(entry) == 0 {
		return errors.New("No valid entry")
	}
	return nil
}

var procfile = flag.String("f", "Procfile", "proc file")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}
	cmd := flag.Args()[0]

	var err error
	switch cmd {
	case "check":
		err = readEntry()
		if err != nil {
			break
		}
		keys := make([]string, len(entry))
		i := 0
		for k := range entry {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		fmt.Printf("valid procfile detected (%s)\n", strings.Join(keys, ", "))
		break
	case "export":
		println("not implemented")
		break
	case "help":
		usage()
		break
	case "run":
		if flag.NArg() != 3 {
			usage()
		}
		err = run(flag.Args()[1], flag.Args()[2])
		break
	case "start":
		err = readEntry()
		if err != nil {
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
		err = start_procs(flag.Args()[1:])
		break
	case "version":
		fmt.Println(version)
		break
	default:
		usage()
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
