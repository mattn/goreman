package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"regexp"
	"sync"
)

// -- process information structure.
type procInfo struct {
	name        string
	environment string
	cmdline     string
	cmd         *exec.Cmd
	port        uint
	setPort     bool
	colorIndex  int

	// True if we called stopProc to kill the process, in which case an
	// *os.ExitError is not the fault of the subprocess
	stoppedBySupervisor bool

	mu      sync.Mutex
	cond    *sync.Cond
	waitErr error
}

var mu sync.Mutex

// process informations named with proc.
var procs []*procInfo

// filename of Procfile.
var procfile = flag.String("f", "Procfile", "proc file")

// base directory
var basedir = flag.String("basedir", "", "base directory")

// true to exit the supervisor
var exitOnError = flag.Bool("exit-on-error", false, "Exit goreman if a subprocess quits with a nonzero return code")

// true to exit the supervisor when all processes stop
var exitOnStop = flag.Bool("exit-on-stop", true, "Exit goreman if all subprocesses stop")

var maxProcNameLength = 0

var re = regexp.MustCompile(`\$([a-zA-Z]+[a-zA-Z0-9_]+)`)

func findProc(name string) *procInfo {
	mu.Lock()
	defer mu.Unlock()

	for _, proc := range procs {
		if proc.name == name {
			return proc
		}
	}
	return nil
}

// command: start. spawn procs.
func start(ctx context.Context, sig <-chan os.Signal, cfg *config) error {
	// Read configuration file
	b, err := os.ReadFile("goreman.yml")
	if err != nil {
		//level.Error(l).Log("msg", "error reading config file", "err", err)
		fmt.Println("err", err)
	}

	var configuration Configuration
	if err := yaml.Unmarshal(b, &configuration); err != nil {
		//level.Error(l).Log("msg", "error unmarshalling config", "err", err)
		fmt.Println("err", err)
	}

	err = readProcfile(configuration)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	// Cancel the RPC server when procs have returned/errored, cancel the
	// context anyway in case of early return.
	defer cancel()
	if len(cfg.Args) > 1 {
		tmp := make([]*procInfo, 0, len(cfg.Args[1:]))
		maxProcNameLength = 0
		for _, v := range cfg.Args[1:] {
			proc := findProc(v)
			if proc == nil {
				return errors.New("unknown proc: " + v)
			}
			tmp = append(tmp, proc)
			if len(v) > maxProcNameLength {
				maxProcNameLength = len(v)
			}
		}
		mu.Lock()
		procs = tmp
		mu.Unlock()
	}
	godotenv.Load()
	procsErr := startProcs(sig, cfg.ExitOnError)
	return procsErr
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
	case "start":
		c := notifyCh()
		err = start(context.Background(), c, cfg)
	default:
		fmt.Println("invalid option")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}
