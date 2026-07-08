package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func exportUpstart(cfg *config, path string) error {
	procfile, err := filepath.Abs(cfg.Procfile)
	if err != nil {
		return err
	}
	// parse .env the same way `goreman start` does (godotenv), so exported
	// values match the runtime environment.
	env, err := godotenv.Read(filepath.Join(filepath.Dir(procfile), ".env"))
	if err != nil {
		env = map[string]string{}
	}

	for _, proc := range procs {
		f, err := os.Create(filepath.Join(path, "app-"+proc.name+".conf"))
		if err != nil {
			return err
		}

		fmt.Fprintf(f, "start on starting app-%s\n", proc.name)
		fmt.Fprintf(f, "stop on stopping app-%s\n", proc.name)
		fmt.Fprintf(f, "respawn\n")
		fmt.Fprintf(f, "\n")

		if proc.setPort {
			fmt.Fprintf(f, "env PORT=%d\n", proc.port)
		}
		for k, v := range env {
			fmt.Fprintf(f, "env %s='%s'\n", k, strings.Replace(v, "'", "\\'", -1))
		}
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "setuid app\n")
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "chdir %s\n", filepath.ToSlash(filepath.Dir(procfile)))
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "exec %s\n", proc.cmdline)

		f.Close()
	}
	return nil
}

// command: export.
func export(cfg *config, format, path string) error {
	err := readProcfile(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	switch format {
	case "upstart":
		return exportUpstart(cfg, path)
	}
	return errors.New("unknown format: " + format)
}
