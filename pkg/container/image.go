package container

import (
	"strings"

	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/tarutils"
)

type ContainerImage struct {
	Index    OCIImageIndex
	Metadata ImageMetadata
	Manifest OCIImageManifest
	Layers   []*Layer
	OciPath  string
}

func ContainerImageFromOCITar(ociPath string) (*ContainerImage, error) {
	_, reader, err := tarutils.GetBlobFromFileByName(ociPath, "index.json")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image index")
		return nil, err
	}
	imageIndex, err := tarutils.ParseJsonReaderIntoInterface[OCIImageIndex](reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse image index into struct")
		return nil, err
	}

	_, reader, err = tarutils.GetBlobFromFileByDigest(ociPath, imageIndex.Manifests[0].Digest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image manifest")
		return nil, err
	}
	manifest, err := tarutils.ParseJsonReaderIntoInterface[OCIImageManifest](reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed parse manifest into struct")
		return nil, err
	}

	_, reader, err = tarutils.GetBlobFromFileByDigest(ociPath, manifest.Config.Digest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image manifest")
		return nil, err
	}
	metadata, err := tarutils.ParseJsonReaderIntoInterface[ImageMetadata](reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed parse manifest into struct")
		return nil, err
	}

	containerImage := ContainerImage{
		Index:    imageIndex,
		Metadata: metadata,
		Manifest: manifest,
		Layers:   make([]*Layer, len(manifest.Layers)),
		OciPath:  ociPath,
	}
	containerImage.buildLayers()
	return &containerImage, nil
}

func (ci *ContainerImage) buildLayers() error {
	for index, digest := range ci.Manifest.Layers {
		ci.Layers[index] = NewLayer(ci.OciPath, digest.Digest, strings.HasSuffix(digest.MediaType, "+gzip"))
	}
	return nil
}
