package runner

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/rs/zerolog/log"
)

//go:embed fs_util_build/*
var fsutil embed.FS

//go:embed os_util_build/*
var osutil embed.FS

//go:embed command_util_build/*
var cmdutil embed.FS

type PythonRunner struct {
	utilImport *template.Template
	exec       string
	fs         embed.FS
}

type TemplateData struct {
	DockerfilePath string
	Image          string
}

func (r *PythonRunner) Run(contextData TemplateData, command string) error {
	workDir, err := r.makeTmpWithFS()
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}
	defer os.RemoveAll(workDir)

	err = r.addFileToTmp(contextData.DockerfilePath, workDir)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}
	contextData.DockerfilePath = "./Dockerfile"

	var buffer bytes.Buffer
	r.utilImport.Execute(&buffer, contextData)

	command = buffer.String() + "\n" + command
	cmd := exec.Command(r.exec, "-c", command)
	cmd.Dir = workDir
	var errorOutput bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &errorOutput
	err = cmd.Run()
	if err != nil {
		log.Error().Err(err).Str("stderr", errorOutput.String()).Send()
		return err
	}
	if len(errorOutput.String()) > 0 {
		log.Error().Err(err).Send()
		return errors.New(errorOutput.String())
	}
	return nil
}

func (r PythonRunner) ToString() string {
	return fmt.Sprintf("Exec: %s with preamble %s", r.exec, r.utilImport.Root.String())
}

func (r PythonRunner) addFileToTmp(srcFilePath, tmpDirPath string) error {
	fileName := filepath.Base(srcFilePath)
	dest := filepath.Join(tmpDirPath, fileName)
	// Note: this is not container friendly
	return os.Link(srcFilePath, dest)
}

// TODO: Clean this up
// This may also profit from caching this/reusing the tmp directories...rules are (for now) not run in parallel so reusing this should save disk space
// See also https://github.com/kluctl/go-embed-python/tree/main/embed_util
// Create a temporary directory with the dependency fs mounted
// THIS EXPECTS THE CALLER TO HANDLE CLEANUP
func (r PythonRunner) makeTmpWithFS() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "embedded")
	if err != nil {
		log.Error().Err(err).Msg("Could not create temporary directory")
		return "", err
	}

	// Walk through the embedded files
	err = fs.WalkDir(r.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Read the file from embed.FS
			data, err := r.fs.ReadFile(path)
			if err != nil {
				return err
			}
			tempFilePath := filepath.Join(tempDir, path)
			log.Debug().Str("filepath", tempFilePath).Send()
			err = os.MkdirAll(filepath.Dir(tempFilePath), os.ModePerm)
			if err != nil {
				return err
			}

			err = os.WriteFile(tempFilePath, data, os.ModePerm)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Could process embedde files")
		return "", err
	}

	log.Debug().Str("tmpDir", tempDir).Msg("Fs mounted to temporary directory")
	return tempDir, nil
}
