package runner

import (
	"fmt"
	"text/template"
)

type Runner interface {
	Run(TemplateData, string, int) error
	RunFix(command string)
	ToString() string
}

var targetScore = map[string]int{
	"command": 0,
	"fs":      1,
	"os":      2,
}

func NewPythonRunner(target string) (Runner, error) {
	runner := &PythonRunner{
		exec:             "python3",
		workingDirectory: GetReferencingWorkingDirectoryInstance(),
	}
	importTemplate := ""
	score, ok := targetScore[target]
	if !ok {
		return nil, fmt.Errorf("Unsupported target: %s! Supported targets are: command, fs, os", target)
	}

	if score >= 0 {
		importTemplate += "from command_util_build import commandutil; command_util = commandutil.setup_from_path('{{ .DockerfilePath }}');"
	}
	if score >= 1 {
		importTemplate += "from fs_util_build import fsutil; fs_util = fsutil.setup('{{ .OciImage }}');"
	}
	if score >= 2 {
		importTemplate += "from os_util_build import osutil; os_util = osutil.setup('{{ .DockerImage }}');"
	}

	var err error
	runner.utilImport, err = template.New("").Parse(importTemplate)
	if err != nil {
		return nil, fmt.Errorf("Import template was unrenderable: %s", err.Error())
	}

	return runner, nil
}
