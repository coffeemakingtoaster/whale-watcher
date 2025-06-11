package fsutil

import (
	"os"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type FsUtils struct {
	// TODO: Implement me
	OCI container.ContainerImage
}

// Setup function used for instantiating util struct
func Setup(ociTarpath string) FsUtils {
	image, err := container.ParseImage(ociTarpath)
	if err != nil {
		panic(err)
	}
	return FsUtils{
		OCI: *image,
	}
}

func (ou FsUtils) Dir_content_count(dirPath string) int {
	res, err := os.ReadDir(dirPath)
	if err != nil {
		return -1
	}
	return len(res)
}

func (ou FsUtils) Name() string {
	return "fs_util"
}

func main() {}
