package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Runner struct {
	utilImport string
	exec       string
}

func (r *Runner) Run(command string) error {
	command = r.utilImport + "\n" + command
	cmd := exec.Command(r.exec, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil

}

func NewPythonRunner(target string) (*Runner, error) {
	runner := &Runner{
		exec: "python3",
	}
	switch target {
	case "command":
		runner.utilImport = "from command_util import *"
	case "fs":
		runner.utilImport = "from fs_util import *"
	case "os":
		runner.utilImport = "from os_util import *"
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported target: %s! Supported targets are command, fs, os", target))
	}
	return runner, nil
}

func (r Runner) ToString() string {
	return fmt.Sprintf("Exec: %s with preamble %s", r.exec, r.utilImport)
}
