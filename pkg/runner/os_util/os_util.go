package osutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
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
			return
		}
	}
	panic("Image name not found!")
}

func (ou *OsUtils) ExecCommand(command string) string {
	ou.load()
	log.Debug().Str("command", command).Msg("Executing command")
	return ou.runCommand([]string{"docker", "run", "--rm", ou.image, command})
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
	return stdOutput.String()
}

func main() {}
