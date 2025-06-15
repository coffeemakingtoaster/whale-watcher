package container

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/layerfs"
)

type Layer struct {
	Digest     string
	tarPath    string
	FileSystem layerfs.LayerFS
}

func NewLayer(ociPath, digest string, isGzip bool) *Layer {
	return &Layer{
		Digest:     digest,
		tarPath:    ociPath,
		FileSystem: layerfs.NewLayerFS(ociPath, digest, isGzip),
	}
}
