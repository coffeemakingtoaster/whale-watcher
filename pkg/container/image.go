package container

import (
	"archive/tar"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

type ContainerImage struct {
}

func ParseImage(ociPath string) (*ContainerImage, error) {
	f, err := os.Open(ociPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := tar.NewReader(f)
	var containerImage ContainerImage
	for true {
		header, err := reader.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		log.Debug().Str("name", header.Name).Str("format", header.Format.String()).Send()
	}
	return &containerImage, nil
}
