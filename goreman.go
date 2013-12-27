package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const version = "0.0.2"

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  goreman check                  # Show entries in Procfile
  goreman help [TASK]            # Show this help
  goreman run COMMAND [ARGS...]  # Run a command (start/stop/restart/list)
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
	port	uint
}

// process informations named with proc.
var procs map[string]*procInfo

// filename of Procfile.
var procfile = flag.String("f", "Procfile", "proc file")
// rpc port number.
var port = flag.Uint("p", defaultPort(), "port")
// base directory
var basedir = flag.String("basedir", "", "base directory")
// base of port numbers for app
var baseport = flag.Uint("b", 5000, "base number of port")

var maxProcNameLength = 0

// read Procfile and parse it.
func readProcfile() error {
	procs = map[string]*procInfo{}
	content, err := ioutil.ReadFile(*procfile)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`\$([a-zA-Z]+[a-zA-Z0-9_])`)
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}
		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		if runtime.GOOS == "windows" {
			v = re.ReplaceAllStringFunc(v, func(s string) string {
				return "%" + s[1:] + "%"
			})
		}
		procs[k] = &procInfo{k, v, false, nil, *baseport}
		*baseport++
		if len(k) > maxProcNameLength {
			maxProcNameLength = len(k)
		}
	}
	if len(procs) == 0 {
		return errors.New("No valid entry")
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
	return 8555
}

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
	godotenv.Load()
	go startServer()
	return startProcs()
}

func main() {
	var err error

	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}
	cmd := flag.Args()[0]

	if *basedir != "" {
		err = os.Chdir(*basedir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goreman: %s\n", err.Error())
			os.Exit(1)
		}
	}

	switch cmd {
	case "check":
		err = check()
		break
	case "help":
		usage()
		break
	case "run":
		if flag.NArg() == 3 {
			cmd, proc := flag.Args()[1], flag.Args()[2]
			err = run(cmd, proc)
		} else if flag.NArg() == 2 {
			cmd := flag.Args()[1]
			err = run(cmd, "")
		} else {
			usage()
		}
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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
