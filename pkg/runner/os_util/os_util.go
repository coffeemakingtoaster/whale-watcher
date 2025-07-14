package osutil

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type OsUtils struct {
	OCI     *container.ContainerImage
	workdir string
}

// Setup function used for instantiating util struct
func Setup(ociTarpath string) OsUtils {
	// Should files be extracted?
	image, err := container.ContainerImageFromOCITar(ociTarpath)
	if err != nil {
		panic(err)
	}
	return OsUtils{
		OCI:     image,
		workdir: filepath.Join(filepath.Dir(ociTarpath), "extracted"),
	}
}

func (ou OsUtils) Name() string {
	return "os_util"
}

func (ou *OsUtils) prepare() {
	err := ou.OCI.ExtractToDir(ou.workdir)
	if err != nil {
		panic(err)
	}
}

func (ou *OsUtils) ExecCommand(command string) string {
	ou.prepare()
	log.Debug().Str("command", command).Msg("Executing command")
	//panic(fmt.Sprintf("%v", ou.OCI.Metadata.Config.Env))
	return ""
}

func main() {}
