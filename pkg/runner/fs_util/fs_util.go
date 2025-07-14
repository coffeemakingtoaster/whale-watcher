package fsutil

import (
	"io"
	"strings"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type FsUtils struct {
	OCI *container.ContainerImage
}

// Setup function used for instantiating util struct
func Setup(ociTarpath string) FsUtils {
	image, err := container.ContainerImageFromOCITar(ociTarpath)
	if err != nil {
		panic(err)
	}
	return FsUtils{
		OCI: image,
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

func (fu *FsUtils) OpenFileAtLayer(filePath string, layerIndex int) []string {
	f, err := fu.OCI.Layers[layerIndex].FileSystem.Open(filePath)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), "\n")
}

func (fu *FsUtils) LookForFile(path string) int {
	index := fu.GetLayerCount() - 1
	for index >= 0 {
		ok, deletion := fu.OCI.Layers[index].FileSystem.HasFile(path)
		if ok {
			if deletion {
				return -1
			}
			return index
		}
		index--
	}
	return -1
}

func (ou FsUtils) Name() string {
	return "fs_util"
}

// THIS IS JANKY!
func (fu *FsUtils) GetInstalledPackages() []string {
	return fu.OCI.GetPackageList()
}

func main() {}
