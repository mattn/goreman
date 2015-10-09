package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const version = "0.0.6"

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  goreman check                     # Show entries in Procfile
  goreman help [TASK]               # Show this help
  goreman run COMMAND [PROCESS...]  # Run a command
                                      (start/stop/restart/list/status)
  goreman start [PROCESS]           # Start the application
  goreman version                   # Display Goreman version

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
	port    uint
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

type config struct {
	Procfile string `yaml:"procfile"`
	Port     uint   `yaml:"port"`
	BaseDir  string `yaml:"basedir"`
	BasePort uint   `yaml:"baseport"`
	Args     []string
}

func readConfig() *config {
	var cfg config

	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	cfg.Procfile = *procfile
	cfg.Port = *port
	cfg.BaseDir = *basedir
	cfg.BasePort = *baseport
	cfg.Args = flag.Args()

	b, err := ioutil.ReadFile(".goreman")
	if err == nil {
		yaml.Unmarshal(b, &cfg)
	}
	return &cfg
}

// read Procfile and parse it.
func readProcfile(cfg *config) error {
	procs = map[string]*procInfo{}
	content, err := ioutil.ReadFile(cfg.Procfile)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`\$([a-zA-Z_]+[a-zA-Z0-9_])`)
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
func check(cfg *config) error {
	err := readProcfile(cfg)
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
func start(cfg *config) error {
	err := readProcfile(cfg)
	if err != nil {
		return err
	}
	if len(cfg.Args) > 1 {
		tmp := map[string]*procInfo{}
		for _, v := range cfg.Args[1:] {
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

	cfg := readConfig()

	if cfg.BaseDir != "" {
		err = os.Chdir(cfg.BaseDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goreman: %s\n", err.Error())
			os.Exit(1)
		}
	}

	cmd := cfg.Args[0]
	switch cmd {
	case "check":
		err = check(cfg)
		break
	case "help":
		usage()
		break
	case "run":
		if len(cfg.Args) == 3 {
			cmd, proc := cfg.Args[1], cfg.Args[2]
			err = run(cmd, proc)
		} else if len(cfg.Args) == 2 {
			cmd := cfg.Args[1]
			err = run(cmd, "")
		} else {
			usage()
		}
		break
	case "start":
		err = start(cfg)
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
