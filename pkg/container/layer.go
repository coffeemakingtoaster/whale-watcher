package container

import (
	"fmt"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/layerfs"
)

type Layer struct {
	Digest     string
	Command    string
	tarPath    string
	FileSystem layerfs.LayerFS
}

func (l *Layer) ToString() string {
	return fmt.Sprintf("[%s](%s) %s", l.Digest, l.tarPath, l.FileSystem.ToString())
}

func NewLayer(ociPath, digest, command string, isGzip bool) *Layer {
	return &Layer{
		Command:    command,
		Digest:     digest,
		tarPath:    ociPath,
		FileSystem: layerfs.NewLayerFS(ociPath, digest, isGzip),
	}
}
