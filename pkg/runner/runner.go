package runner

import (
	"errors"
	"fmt"
)

type Runner interface {
	Run(string) error
	ToString() string
}

func NewPythonRunner(target string) (Runner, error) {
	runner := &PythonRunner{
		exec: "python3",
	}
	switch target {
	case "command":
		runner.utilImport = "from command_util_build.commandutil import *"
		runner.fs = cmdutil
	case "fs":
		runner.utilImport = "from fs_util_build.fsutil import *"
		runner.fs = fsutil
	case "os":
		runner.utilImport = "from os_util_build.osutil import *"
		runner.fs = osutil
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported target: %s! Supported targets are: command, fs, os", target))
	}
	return runner, nil
}
