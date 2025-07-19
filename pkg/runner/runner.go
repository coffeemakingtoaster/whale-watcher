package runner

import (
	"fmt"
	"text/template"
)

type Runner interface {
	Run(TemplateData, string) error
	RunFix(command string)
	ToString() string
}

func NewPythonRunner(target string) (Runner, error) {
	runner := &PythonRunner{
		exec:             "python3",
		workingDirectory: GetReferencingWorkingDirectoryInstance(),
	}
	importTemplate := ""
	switch target {
	case "command":
		importTemplate = "from command_util_build import commandutil; command_util = commandutil.setup_from_path('{{ .DockerfilePath }}')"
	case "fs":
		importTemplate = "from fs_util_build import fsutil; fs_util = fsutil.setup('{{ .OciImage }}')"
	case "os":
		importTemplate = "from os_util_build import osutil; os_util = osutil.setup('{{ .DockerImage }}')"
	default:
		return nil, fmt.Errorf("Unsupported target: %s! Supported targets are: command, fs, os", target)
	}
	var err error
	runner.utilImport, err = template.New("").Parse(importTemplate)
	if err != nil {
		return nil, fmt.Errorf("Import template was unrenderable: %s", err.Error())
	}

	return runner, nil
}
