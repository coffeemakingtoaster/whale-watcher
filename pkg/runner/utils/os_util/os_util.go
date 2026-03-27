package osutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type OsUtils struct {
	workdir       string
	loaded        bool
	image         string
	dockerTarPath string
}

// Setup function used for instantiating util struct
func Setup(dockerTarpath string) OsUtils {
	// Should files be extracted?
	return OsUtils{
		workdir:       filepath.Dir(dockerTarpath),
		loaded:        false,
		image:         "",
		dockerTarPath: dockerTarpath,
	}
}

func (ou OsUtils) Name() string {
	return "os_util"
}

func (ou *OsUtils) load() {
	if ou.loaded {
		return
	}
	output := ou.runCommand([]string{"docker", "load", "-i", ou.dockerTarPath})

	lines := strings.Split(output, "\n")
	for i := range lines {
		if strings.Contains(lines[i], "Loaded image:") {
			image := strings.Replace(lines[i], "Loaded image:", "", 1)
			image = strings.TrimSpace(image)
			ou.image = image
			ou.loaded = true
			break
		}
	}
	if !ou.loaded {
		panic("image could not be loaded")
	}
}

// this returns a json string, parsing json must be done in the ruleset as of now
// TODO: I need to solve struct/map -> dict conversion somehow
func (ou *OsUtils) GetImageMetadata() string {
	ou.load()
	return ou.runCommand([]string{"docker", "image", "inspect", "--format=json", ou.image})
}

func (ou *OsUtils) ExecCommand(command string) string {
	ou.load()
	return ou.runCommand([]string{"docker", "run", "--entrypoint", "/bin/sh", "--rm", ou.image, "-c", command})
}

func (ou *OsUtils) runCommand(command []string) string {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = ou.workdir

	var errorOutput bytes.Buffer
	var stdOutput bytes.Buffer

	cmd.Stdout = &stdOutput
	cmd.Stderr = &errorOutput

	err := cmd.Run()
	if err != nil {
		fmt.Println(errorOutput.String())
		panic(err)
	}
	return strings.TrimSpace(stdOutput.String())
}

func main() {}
