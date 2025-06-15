package fsutil

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type FsUtils struct {
	OCI container.ContainerImage
}

// Setup function used for instantiating util struct
func Setup(ociTarpath string) FsUtils {
	image, err := container.ContainerImageFromOCITar(ociTarpath)
	if err != nil {
		panic(err)
	}
	return FsUtils{
		OCI: *image,
	}
}

func (fu *FsUtils) Dir_content_count(dirPath string) int {
	files := fu.OCI.Layers[len(fu.OCI.Layers)-1].FileSystem.Ls(dirPath)
	return len(files)
}

func (fu *FsUtils) Ls_Layer(dirPath string, layerIndex int) []string {
	return fu.OCI.Layers[layerIndex].FileSystem.Ls(dirPath)
}

func (ou FsUtils) Name() string {
	return "fs_util"
}

func main() {}
