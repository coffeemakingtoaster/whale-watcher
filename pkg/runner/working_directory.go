package runner

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	COMMAND_UTIL_LEVEL = iota
	FS_UTIL_LEVEL
	OS_UTIL_LEVEL
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
	tmpDirPath         string
	refCount           int
	isPopulated        bool
	current_util_level int
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

func (rwd *RunnerWorkingDirectory) Free() bool {
	rwd.refCount--
	if rwd.refCount > 0 {
		log.Debug().Msgf("Free was called for working directory but ref count has not hit 0")
		return false
	}
	err := os.RemoveAll(rwd.tmpDirPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to cleanup working directory for runner at %s", rwd.tmpDirPath)
		return false
	}
	log.Debug().Msgf("Working directory cleaned up (ref count was 0)")
	instance = nil
	return true
}

func (rwd *RunnerWorkingDirectory) Populate(dockerFilePath, ociImagePath, dockerImagePath string, util_level int) {
	var err error
	err = rwd.extractUtils(util_level)
	if err != nil {
		log.Warn().Err(err).Msg("Error preparing utils")
	}
	if rwd.isPopulated {
		return
	}
	err = addFileToWorkingDirectory(dockerFilePath, rwd.tmpDirPath, "Dockerfile")
	if err != nil {
		log.Warn().Err(err).Msgf("Could not add %s to working directory %s", dockerFilePath, rwd.tmpDirPath)
		return
	}
	if !config.AllowsTarget("fs") && !config.AllowsTarget("os") {
		log.Info().Msg("Not adding container artifacts to working directory as they are not needed for allowed targets")
	} else {
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

func (rwd *RunnerWorkingDirectory) extractUtils(utilLevel int) error {
	// These writes here could fail when using this concurrently
	var err error
	if utilLevel >= COMMAND_UTIL_LEVEL && rwd.current_util_level < COMMAND_UTIL_LEVEL {
		err = unpackFsToDir(cmdutil, rwd.tmpDirPath)
		if err != nil {
			return err
		}
		rwd.current_util_level = COMMAND_UTIL_LEVEL
	}

	if utilLevel >= FS_UTIL_LEVEL && rwd.current_util_level < FS_UTIL_LEVEL {
		err = unpackFsToDir(fsutil, rwd.tmpDirPath)
		if err != nil {
			return err
		}
		rwd.current_util_level = FS_UTIL_LEVEL
	}

	if utilLevel >= OS_UTIL_LEVEL && rwd.current_util_level < OS_UTIL_LEVEL {
		err = unpackFsToDir(osutil, rwd.tmpDirPath)
		if err != nil {
			return err
		}
		rwd.current_util_level = OS_UTIL_LEVEL
	}

	// no fix utils needed if we are running again or in nofix
	if viper.GetBool("nofix") || rwd.isPopulated {
		return nil
	}

	return unpackFsToDir(fixutil, rwd.tmpDirPath)
}

func newRunnerWorkingDirectory() (*RunnerWorkingDirectory, error) {
	dirPath, err := getTmpDir()
	if err != nil {
		return nil, err
	}

	return &RunnerWorkingDirectory{
		tmpDirPath:         dirPath,
		refCount:           0,
		current_util_level: -1,
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

func (rwd *RunnerWorkingDirectory) ForceFree() {
	log.Warn().Int("Dangling references", rwd.refCount).Msg("Forced working directory cleanup! This likely indicated that something went (very) wrong.")
	err := os.RemoveAll(rwd.tmpDirPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to cleanup working directory for runner at %s", rwd.tmpDirPath)
		return
	}
	instance = nil
}
