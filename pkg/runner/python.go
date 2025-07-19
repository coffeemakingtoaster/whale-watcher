package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/rs/zerolog/log"
)

type PythonRunner struct {
	utilImport       *template.Template
	exec             string
	workingDirectory *RunnerWorkingDirectory
}

type TemplateData struct {
	DockerfilePath string
	OciImage       string
	DockerImage    string
}

func (r *PythonRunner) RunFix(command string) {
	log.Info().Msg("Running fix")
	w := GetReferencingWorkingDirectoryInstance()
	defer w.Free()
	importTemplate := "from command_util_build import commandutil; command_util = commandutil.setup_from_path('{{ .DockerfilePath }}');from fix_util_build import fixutil; fix_util = fixutil.setup_from_path('{{ .DockerfilePath }}');"

	contextData := TemplateData{
		DockerfilePath: "./Dockerfile",
		OciImage:       "./out.tar",
		DockerImage:    "./out_docker.tar",
	}

	tpl, _ := template.New("").Parse(importTemplate)

	var buffer bytes.Buffer
	tpl.Execute(&buffer, contextData)

	command = buffer.String() + "\n" + command
	cmd := exec.Command(r.exec, "-c", command)
	cmd.Dir = r.workingDirectory.tmpDirPath

	var errorOutput bytes.Buffer
	var stdOutput bytes.Buffer

	cmd.Stdout = &stdOutput
	cmd.Stderr = &errorOutput

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Str("stderr", errorOutput.String()).Str("stdout", stdOutput.String()).Send()
	}
}

func (r *PythonRunner) Run(contextData TemplateData, command string) error {

	defer r.workingDirectory.Free()

	r.workingDirectory.Populate(contextData.DockerfilePath, contextData.OciImage, contextData.DockerImage)

	contextData.DockerfilePath = "./Dockerfile"
	contextData.OciImage = "./out.tar"
	contextData.DockerImage = "./out_docker.tar"

	var buffer bytes.Buffer
	r.utilImport.Execute(&buffer, contextData)

	command = buffer.String() + "\n" + command
	cmd := exec.Command(r.exec, "-c", command)
	cmd.Dir = r.workingDirectory.tmpDirPath

	var errorOutput bytes.Buffer
	var stdOutput bytes.Buffer

	cmd.Stdout = &stdOutput
	cmd.Stderr = &errorOutput

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Str("stderr", errorOutput.String()).Str("stdout", stdOutput.String()).Send()
		return err
	}

	return nil
}

func (r PythonRunner) ToString() string {
	return fmt.Sprintf("Exec: %s with preamble %s", r.exec, r.utilImport.Root.String())
}

func (r PythonRunner) addFileToTmp(srcFilePath, tmpDirPath string) error {
	fileName := filepath.Base(srcFilePath)
	dest := filepath.Join(tmpDirPath, fileName)
	log.Debug().Str("filepath", dest).Send()

	// Note: this is not container friendly
	return os.Link(srcFilePath, dest)
}
