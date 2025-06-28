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
	Image          string
}

func (r *PythonRunner) Run(contextData TemplateData, command string) error {

	defer r.workingDirectory.Free()

	r.workingDirectory.Populate(contextData.DockerfilePath, contextData.Image)

	contextData.DockerfilePath = "./Dockerfile"
	contextData.Image = "./out.tar"

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
