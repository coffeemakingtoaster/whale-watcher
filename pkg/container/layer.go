package container

import (
	"fmt"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/layerfs"
)

type Layer struct {
	Digest     string
	tarPath    string
	FileSystem layerfs.LayerFS
}

func (l *Layer) ToString() string {
	return fmt.Sprintf("[%s](%s) %s", l.Digest, l.tarPath, l.FileSystem.ToString())
}

func NewLayer(ociPath, digest string, isGzip bool) *Layer {
	return &Layer{
		Digest:     digest,
		tarPath:    ociPath,
		FileSystem: layerfs.NewLayerFS(ociPath, digest, isGzip),
	}
}
