package runner

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

//go:embed _fs_util_build/*
var fsutil embed.FS

//go:embed _os_util_build/*
var osutil embed.FS

//go:embed _command_util_build/*
var cmdutil embed.FS

//go:embed _fix_util_build/*
var fixutil embed.FS

var lock = &sync.Mutex{}

type RunnerWorkingDirectory struct {
	tmpDirPath  string
	refCount    int
	isPopulated bool
}

var instance *RunnerWorkingDirectory

func GetReferencingWorkingDirectoryInstance() *RunnerWorkingDirectory {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		var err error
		instance, err = newRunnerWorkingDirectory()
		if err != nil {
			log.Error().Err(err).Msg("Could not instantiate tmp directory for python runner environment")
			return nil
		}

	}
	instance.refCount++
	return instance
}

func (rwd *RunnerWorkingDirectory) GetAbsolutePath(path string) string {
	return filepath.Join(rwd.tmpDirPath, path)
}

func (rwd *RunnerWorkingDirectory) Free() {
	rwd.refCount--
	if rwd.refCount > 0 {
		log.Debug().Msgf("Free was called for working directory but ref count has not hit 0")
		return
	}
	err := os.RemoveAll(rwd.tmpDirPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to cleanup working directory for runner at %s", rwd.tmpDirPath)
		return
	}
	log.Debug().Msgf("Working directory cleaned up (ref count was 0)")
	instance = nil
}

func (rwd *RunnerWorkingDirectory) Populate(dockerFilePath, ociImagePath, dockerImagePath string) {
	if rwd.isPopulated {
		return
	}
	var err error
	err = addFileToWorkingDirectory(dockerFilePath, rwd.tmpDirPath, "Dockerfile")
	if err != nil {
		log.Warn().Err(err).Msgf("Could not add %s to working directory %s", dockerFilePath, rwd.tmpDirPath)
		return
	}
	err = addFileToWorkingDirectory(ociImagePath, rwd.tmpDirPath, "out.tar")
	if err != nil {
		log.Warn().Err(err).Msgf("Could not add %s to working directory %s", ociImagePath, rwd.tmpDirPath)
		return
	}
	err = addFileToWorkingDirectory(dockerImagePath, rwd.tmpDirPath, "out_docker.tar")
	if err != nil {
		log.Warn().Err(err).Msgf("Could not add %s to working directory %s", ociImagePath, rwd.tmpDirPath)
		return
	}
	rwd.isPopulated = true
}

func addFileToWorkingDirectory(source, workingDirectory, newName string) error {
	destination := filepath.Join(workingDirectory, newName)
	log.Debug().Str("source", source).Str("dest", destination).Send()

	err := os.Link(source, destination)
	// Assume that linking failed due to cross device things
	// copy should work
	if err != nil {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		// Write data to dst
		err = os.WriteFile(destination, data, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func newRunnerWorkingDirectory() (*RunnerWorkingDirectory, error) {
	dirPath, err := getTmpDir()
	if err != nil {
		return nil, err
	}
	err = unpackFsToDir(osutil, dirPath)
	err = unpackFsToDir(fsutil, dirPath)
	err = unpackFsToDir(cmdutil, dirPath)
	err = unpackFsToDir(fixutil, dirPath)

	if err != nil {
		return nil, err
	}

	return &RunnerWorkingDirectory{
		tmpDirPath: dirPath,
		refCount:   0,
	}, nil
}

func getTmpDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "embedded")
	if err != nil {
		log.Error().Err(err).Msg("Could not create temporary directory")
		return "", err
	}
	return tempDir, nil
}

// TODO: Clean this up
// This may also profit from caching this/reusing the tmp directories...rules are (for now) not run in parallel so reusing this should save disk space
// See also https://github.com/kluctl/go-embed-python/tree/main/embed_util
// Create a temporary directory with the dependency fs mounted
// THIS EXPECTS THE CALLER TO HANDLE CLEANUP
func unpackFsToDir(toUnpack embed.FS, dirPath string) error {
	// Walk through the embedded files
	err := fs.WalkDir(toUnpack, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Read the file from embed.FS
			data, err := toUnpack.ReadFile(path)
			if err != nil {
				return err
			}
			// Directories start with _<target>_util_build
			// This ensures that go utilities ignore the build directories but adds a layer of confusion here
			path = strings.TrimPrefix(path, "_")
			tempFilePath := filepath.Join(dirPath, path)
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
		return err
	}

	log.Debug().Str("tmpDir", dirPath).Msg("Fs mounted to temporary directory")
	return nil
}
