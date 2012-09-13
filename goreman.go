package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

func getEntry() (map[string]string, error) {
	content, err := ioutil.ReadFile("Procfile")
	if err != nil {
		return nil, err
	}
	entry := map[string]string {}
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) == 2 && tokens[0][0] != '#' {
			entry[strings.TrimSpace(tokens[0])] = strings.TrimSpace(tokens[1])
		}
	}
	return entry, nil
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

func main() {
	if len(os.Args) == 1 {
		usage();
	}
	cmd := os.Args[1]

	pidfile := filepath.Join(os.TempDir(), "goreman.pid")
	var err error
	var pidf *os.File
	var b []byte
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
		b, err = ioutil.ReadFile(pidfile)
		if err != nil {
			break
		}
		println(string(b))
		println("not implemented")
		break
	case "start":
		pidf, err = os.Create(pidfile)
		if err != nil {
			break
		}
		fmt.Fprintf(pidf, "%d", syscall.Getpid())
		defer func() {
			pidf.Close()
			syscall.Unlink(pidfile)
		}()
		err = start(os.Args[2:])
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
