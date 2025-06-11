package runner

import (
	"errors"
	"fmt"
	"text/template"
)

type Runner interface {
	Run(TemplateData, string) error
	ToString() string
}

func NewPythonRunner(target string) (Runner, error) {
	runner := &PythonRunner{
		exec: "python3",
	}
	importTemplate := ""
	switch target {
	case "command":
		importTemplate = "from command_util_build import commandutil; command_util = commandutil.setup_from_path('{{ .DockerfilePath }}')"
		runner.fs = cmdutil
	case "fs":
		importTemplate = "from fs_util_build import fsutil; fs_util = fsutil.setup('{{ .Image }}')"
		runner.fs = fsutil
	case "os":
		importTemplate = "from os_util_build import osutil; os_util = osutil.setup()"
		runner.fs = osutil
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported target: %s! Supported targets are: command, fs, os", target))
	}
	var err error
	runner.utilImport, err = template.New("").Parse(importTemplate)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Import template was unrenderable: %s", err.Error()))
	}

	return runner, nil
}
