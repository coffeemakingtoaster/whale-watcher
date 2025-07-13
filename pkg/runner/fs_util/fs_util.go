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

func (fu *FsUtils) GetLayerCount() int {
	return len(fu.OCI.Layers)
}

func (fu *FsUtils) DirContentCount(dirPath string) int {
	files := fu.OCI.Layers[len(fu.OCI.Layers)-1].FileSystem.Ls(dirPath)
	return len(files)
}

func (fu *FsUtils) LsLayer(dirPath string, layerIndex int) []string {
	return fu.OCI.Layers[layerIndex].FileSystem.Ls(dirPath)
}

func (FsUtils) Name() string {
	return "fs_util"
}

func (fu *FsUtils) GetBaseImageIdentifier() string {
	return ""
}

// THIS IS JANKY!
func (fu *FsUtils) GetInstalledPackages() []string {
	return []string{}
}

func main() {}
