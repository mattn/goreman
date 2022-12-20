package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"text/template/parse"
)

type config struct {
	Procfile string `yaml:"procfile"`
	BaseDir  string `yaml:"basedir"`
	BasePort uint   `yaml:"baseport"`
	Args     []string
	// If true, exit the supervisor process if a subprocess exits with an error.
	ExitOnError bool `yaml:"exit_on_error"`
}

// Service is a single configuration option for a service we want to run
type Service struct {
	Command     string `yaml:"command"`
	Environment string `yaml:"environment"`
	Enable      bool   `yaml:"enable"`
	// Variables are string mappings, the key can be used as $KEY in the "Command" string. It will be interpolated when
	// it is used to spawn the proc
	Variables []map[string]string `yaml:"variables"`
}

// Configuration holds a configuration, the key of the map is the name of the configuration. This is a string defined by
// the user to differentiate the various services started.
type Configuration map[string]Service

func readConfig() *config {
	var cfg config

	flag.Parse()

	cfg.Procfile = *procfile
	cfg.BaseDir = *basedir
	cfg.BasePort = *baseport
	cfg.ExitOnError = *exitOnError
	cfg.Args = flag.Args()
	return &cfg
}

// read Procfile and parse it.
func readProcfile(cfg Configuration) error {
	mu.Lock()
	defer mu.Unlock()

	procs = []*procInfo{}
	index := 0
	for key, service := range cfg {
		// Skip all the services that don't pass the validation (Not enabled, erroneous configuration etc.)
		if !service.Valid() {
			continue
		}
		// Create proc based on configuration
		cmd, err := service.InterpolatedCommand()
		if err != nil {
			return err
		}
		proc := &procInfo{
			name:        fmt.Sprintf("%s-%s", key, service.Environment),
			environment: service.Environment,
			cmdline:     cmd,
			colorIndex:  index,
		}
		proc.cond = sync.NewCond(&proc.mu)
		procs = append(procs, proc)
		index = (index + 1) % len(colors)
	}
	if len(procs) == 0 {
		return errors.New("no valid service entry in configuration file")
	}
	return nil
}

func (s Service) InterpolatedCommand() (string, error) {
	var finalCommand string
	tmpl, err := template.New("command").Parse(s.Command)
	if err != nil {
		return "", err
	}

	// Replace variables in command string if variables exist, otherwise we just return the original command
	variables := ListTemplateFields(tmpl)
	if len(s.Variables) > 0 {
		for _, val := range s.Variables {
			for _, val := range val {
				for _, variable := range variables {
					if strings.Contains(s.Command, variable) {
						finalCommand = strings.Replace(s.Command, variable, val, -1)
					}
				}
			}
		}
		return finalCommand, nil
	}
	return s.Command, nil
}

// Valid returns true if a service is enabled and has all the required values set
func (s Service) Valid() bool {
	// Fail early if the service is not enabled
	if !s.Enable {
		return false
	}

	vars, err := extractVariables(s.Command)
	if err != nil {
		return false
	}

	// Fail early if different counts
	if len(vars) != len(s.Variables) {
		return false
	}

	vm := make(map[string]struct{})
	for _, v := range vars {
		if _, ok := vm[v]; !ok {
			vm[v] = struct{}{}
		}
	}
	for _, variable := range s.Variables {
		for key, _ := range variable {
			if val, ok := vm[key]; ok {
				if vm[key] != val {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}

// extractVariables parses a command template and returns the Go template variables that were used
func extractVariables(command string) ([]string, error) {
	tmpl, err := template.New("command").Parse(command)
	if err != nil {
		return nil, err
	}
	variables := ListTemplateFields(tmpl)
	for i, _ := range variables {
		variables[i] = strings.Replace(variables[i], "{{", "", -1)
		variables[i] = strings.Replace(variables[i], "}}", "", -1)
		variables[i] = strings.Replace(variables[i], ".", "", -1)
		variables[i] = strings.ToLower(variables[i])
	}
	return variables, nil
}

// ListTemplateFields lists the fields used in a template. Sourced and adapted from: https://stackoverflow.com/a/40584967
func ListTemplateFields(t *template.Template) []string {
	return listNodeFields(t.Tree.Root, nil)
}

// listNodeFields iterates over the parsed tree and extracts fields
func listNodeFields(node parse.Node, res []string) []string {
	//fmt.Println("p", node.String())
	//fmt.Println("p", node.Type())
	// Only looking at fields, needs to be adapted if further template entities should be supported
	//if node.Type() == parse.NodeField {
	//	res = append(res, node.String())
	//}

	if node.Type() == parse.NodeAction {
		res = append(res, node.String())
	}

	if ln, ok := node.(*parse.ListNode); ok {
		for _, n := range ln.Nodes {
			res = listNodeFields(n, res)
		}
	}
	return res
}

func (c Configuration) InterpolatedCommands() []string {
	var commands []string
	for _, option := range c {
		var finalCommand string
		// Replace variables in command string if variables exist, otherwise we just return the original command
		if len(option.Variables) > 0 {
			for _, val := range option.Variables {
				for key, val := range val {
					if strings.Contains(option.Command, "$"+strings.ToUpper(key)) {
						finalCommand = strings.Replace(option.Command, "$"+strings.ToUpper(key), val, -1)
					}
				}
			}
			commands = append(commands, finalCommand)
		} else {
			commands = append(commands, option.Command)
		}
	}
	return commands
}
