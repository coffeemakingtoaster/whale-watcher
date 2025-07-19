package fetcher

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/config"
)

func FetchContainerFiles() (string, string) {
	var dockerfilePath string
	var ociPath string
	var err error

	cfg := config.GetConfig()
	if cfg.Target.RepositoryURL == "" {
		dockerfilePath = cfg.Target.DockerfilePath
	} else {
		dockerfilePath, err = loadDockerfileFromRepository(cfg.Target.RepositoryURL, cfg.Target.Branch, cfg.Target.DockerfilePath)
		if err != nil {
			log.Warn().Err(err).Msg("Could not load dockerfile from repository")
		}
	}

	if cfg.Target.Image == "" {
		ociPath = cfg.Target.OciPath
	} else {
		ociPath, err = loadImageFromRegistry(cfg.Target.Image)
		if err != nil {
			log.Warn().Err(err).Msg("Could not load image from repository")
		}

	}

	return dockerfilePath, ociPath
}

func loadImageFromRegistry(image string) (string, error) {
	log.Info().Str("image", image).Msg("Downloading image from registry")
	tmpDirPath, err := os.MkdirTemp("", "filecache")
	if err != nil {
		if !os.IsExist(err) {
			return "", err
		}
	}

	destination := filepath.Join(tmpDirPath, "image.tar")
	err = LoadTarToPath(image, destination)
	if err != nil {
		return "", err
	}

	log.Info().Str("image", image).Msg("Successful download")
	return destination, nil
}

func LoadTarToPath(image, destination string) error {
	ref, err := name.ParseReference(image)
	if err != nil {
		return err
	}

	img, err := remote.Image(ref)
	if err != nil {
		return err
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}

	err = tarball.Write(ref, img, file)
	if err != nil {
		log.Error().Err(err).Msgf("Could not download image %s", image)
		return err
	}
	return nil
}

func loadDockerfileFromRepository(repositoryURL, branch, dockerfilePath string) (string, error) {
	data, err := getFileFromRepository(repositoryURL, branch, dockerfilePath)
	if err != nil {
		log.Error().Msg("Could not retrieve filedata from repository")
		return "", err
	}
	tmpDirPath, err := os.MkdirTemp("", "filecache")
	if err != nil {
		log.Error().Msg("Could not create temporary directory")
		return "", err
	}
	loadedPath := filepath.Join(tmpDirPath, "Dockerfile")
	err = os.WriteFile(loadedPath, data, 0755)
	if err != nil {
		log.Error().Msg("Could not write repository file data to tmp directory")
		return "", err
	}
	return loadedPath, nil
}

func getFileFromRepository(repositoryURL, branch, path string) ([]byte, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	repository, err := git.Clone(
		storer,
		fs,
		&git.CloneOptions{URL: repositoryURL},
	)

	if err != nil {
		return []byte{}, err
	}
	w, err := repository.Worktree()
	if err != nil {
		return []byte{}, err
	}
	dockerfileDirectory := filepath.Dir(path)

	branchReference := plumbing.NewBranchReferenceName(branch)

	err = w.Checkout(&git.CheckoutOptions{SparseCheckoutDirectories: []string{dockerfileDirectory}, Branch: plumbing.ReferenceName(branchReference)})
	if err != nil {
		log.Error().Msg("Could not checkout")
		return []byte{}, err
	}

	fileHandle, err := w.Filesystem.Open(path)
	if err != nil {
		log.Error().Str("path", path).Msg("Could not open file in worktree")
		return []byte{}, err
	}

	var data bytes.Buffer
	_, err = io.Copy(&data, fileHandle)
	if err != nil {
		log.Error().Msg("Could not copy file data to buffer")
		return []byte{}, err
	}
	return data.Bytes(), nil
}
