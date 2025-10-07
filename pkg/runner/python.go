package runner

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//go:embed python.tmpl
var pythonTemplate string

type PythonRunner struct {
	exec             string
	workingDirectory *RunnerWorkingDirectory
}

type TemplateData struct {
	DockerfilePath string
	OciImage       string
	DockerImage    string
	Rules          []*rules.Rule
	NoFix          bool
	HighestTarget  string
}

func (r *PythonRunner) Run(ruleSet rules.RuleSet, ociTarPath, dockerFilepath, dockerTarPath string) (map[string]RunnerResult, error) {
	var err error

	defer r.workingDirectory.Free()

	r.workingDirectory.Populate(dockerFilepath, ociTarPath, dockerTarPath, ruleSet.GetHighestTarget())
	contextData := TemplateData{
		DockerfilePath: "./Dockerfile",
		OciImage:       "./out.tar",
		DockerImage:    "./out_docker.tar",
		Rules:          ruleSet.Rules,
		NoFix:          viper.GetBool("no_fix"),
		HighestTarget:  ruleSet.GetHighestTarget(),
	}

	tpl, err := template.New("pythonExecutionContent").Parse(pythonTemplate)

	if err != nil {
		return map[string]RunnerResult{}, err
	}

	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, contextData)
	if err != nil {
		return map[string]RunnerResult{}, err
	}

	err = writeToFile(r.workingDirectory.GetAbsolutePath("ww.py"), buffer)
	if err != nil {
		return map[string]RunnerResult{}, err
	}

	cmd := exec.Command(r.exec, "ww.py")
	cmd.Dir = r.workingDirectory.tmpDirPath

	var errorOutput bytes.Buffer
	var stdOutput bytes.Buffer

	cmd.Stdout = &stdOutput
	cmd.Stderr = &errorOutput

	err = cmd.Run()
	if err != nil {
		log.Error().Str("stderr", errorOutput.String()).Str("stdout", stdOutput.String()).Send()
		// signal aborted indicates an issue with the gopy build result, advancing is useless
		if strings.Contains(err.Error(), "signal: aborted (core dumped)") {
			panic(err)
		}
		return map[string]RunnerResult{}, err
	}
	return r.parseOutput(stdOutput.String())
}

func writeToFile(p string, data bytes.Buffer) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(data.String())
	return nil
}

func (r *PythonRunner) parseOutput(stdOut string) (map[string]RunnerResult, error) {

	result := make(map[string]RunnerResult)

	lines := strings.Split(stdOut, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fields := strings.Split(line, "\t")
		var key, status, autofix string

		if len(fields) >= 3 {
			key = strings.Trim(fields[0][len("WW_KEY="):], "'")
			status = strings.TrimPrefix(fields[1], "WW_STATUS=")
			autofix = strings.TrimPrefix(fields[2], "WW_AUTOFIX=")
		}

		if len(status) == 0 || len(key) == 0 {
			return result, fmt.Errorf("Cannot parse line: %s", line)
		}

		result[key] = RunnerResult{
			Autofix: autofix == "True",
			Success: status == "True",
		}
	}
	return result, nil
}
