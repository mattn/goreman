package main

import (
	"flag"
	"fmt"
	"errors"
	"os/exec"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

const version = "0.0.1"

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  goreman check                  # Show entries in Procfile
  goreman help [TASK]            # Show this help
  goreman run COMMAND [ARGS...]  # Run a command (start/stop/restart)
  goreman start [PROCESS]        # Start the application
  goreman version                # Display Goreman version

Options:
  -f # Default: Procfile
`)
	os.Exit(0)
}

type proc_info struct {
	proc string
	cmdline string
	quit bool
	cmd *exec.Cmd
}
var procs map[string]*proc_info

func read_procfile() error {
	procs = map[string]*proc_info {}
	content, err := ioutil.ReadFile(*procfile)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) == 2 && tokens[0][0] != '#' {
			k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
			procs[k] = &proc_info{k, v, false, nil}
		}
	}
	if len(procs) == 0 {
		return errors.New("No valid entry")
	}
	return nil
}

var procfile = flag.String("f", "Procfile", "proc file")

func check() error {
	err := read_procfile()
	if err != nil {
		return err
	}
	keys := make([]string, len(procs))
	i := 0
	for k := range procs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	fmt.Printf("valid procfile detected (%s)\n", strings.Join(keys, ", "))
	return nil
}

func start() error {
	err := read_procfile()
	if err != nil {
		return err
	}
	if flag.NArg() > 1 {
		tmp := map[string]*proc_info {}
		for _, v := range flag.Args()[1:] {
			tmp[v] = procs[v]
		}
		procs = tmp
	}
	go start_server()
	return start_procs()
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}
	cmd := flag.Args()[0]

	var err error
	switch cmd {
	case "check":
		err = check()
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
		cmd, proc := flag.Args()[1], flag.Args()[2]
		err = run(cmd, proc)
		break
	case "start":
		err = start()
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
