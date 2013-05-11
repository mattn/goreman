package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const version = "0.0.2"

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  goreman check                  # Show entries in Procfile
  goreman help [TASK]            # Show this help
  goreman run COMMAND [ARGS...]  # Run a command (start/stop/restart)
  goreman start [PROCESS]        # Start the application
  goreman version                # Display Goreman version

Options:
`)
	flag.PrintDefaults()
	os.Exit(0)
}

// -- process information structure.
type procInfo struct {
	proc    string
	cmdline string
	quit    bool
	cmd     *exec.Cmd
}

// process informations named with proc.
var procs map[string]*procInfo

// filename of Procfile.
var procfile = flag.String("f", "Procfile", "proc file")

// read Procfile and parse it.
func readProcfile() error {
	procs = map[string]*procInfo{}
	content, err := ioutil.ReadFile(*procfile)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) == 2 && tokens[0][0] != '#' {
			k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
			procs[k] = &procInfo{k, v, false, nil}
		}
	}
	if len(procs) == 0 {
		return errors.New("No valid entry")
	}
	return nil
}

// read .env file (if exists) and set environments.
func readEnvfile() error {
	content, err := ioutil.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err){
			return nil
		}
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) == 2 && tokens[0][0] != '#' {
			k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
			os.Setenv(k, v)
		}
	}
	return nil
}

// default port
func defaultPort() uint {
	s := os.Getenv("GOREMAN_RPC_PORT")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err == nil {
			return uint(i)
		}
	}
	return 5555
}

// rpc port number.
var port = flag.Uint("p", defaultPort(), "port")

// command: check. show Procfile entries.
func check() error {
	err := readProcfile()
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

// command: start. spawn procs.
func start() error {
	err := readProcfile()
	if err != nil {
		return err
	}
	if flag.NArg() > 1 {
		tmp := map[string]*procInfo{}
		for _, v := range flag.Args()[1:] {
			tmp[v] = procs[v]
		}
		procs = tmp
	}
	err = readEnvfile()
	if err != nil {
		return err
	}
	go startServer()
	return startProcs()
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
