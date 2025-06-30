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
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/config"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/fetcher/ghcr"
)

func FetchContainerFiles() (string, string) {
	var dockerfilePath string
	var ociPath string

	cfg := config.GetConfig()
	if cfg.Target.RepositoryURL == "" {
		dockerfilePath = cfg.Target.DockerfilePath
	} else {
		dockerfilePath, _ = loadDockerfileFromRepository(cfg.Target.RepositoryURL, cfg.Target.Branch, cfg.Target.DockerfilePath)
	}

	if cfg.Target.Image == "" {
		ociPath = cfg.Target.OciPath
	} else {
		ociPath, _ = loadImageFromRegistry(cfg.Target.Image)
	}

	return dockerfilePath, ociPath
}

func loadImageFromRegistry(image string) (string, error) {
	tmpDirPath, err := os.MkdirTemp("", "filecache")
	if err != nil {
		if !os.IsExist(err) {
			return "", err
		}
	}
	err = ghcr.DownloadOciToPath(image, filepath.Join(tmpDirPath, "image.tar"))
	if err != nil {
		log.Error().Err(err).Msgf("Could not download image %s", image)
		return "", err
	}
	return tmpDirPath, err
}

func loadDockerfileFromRepository(repositoryURL, branch, dockerfilePath string) (string, error) {
	data, err := getFileFromRepository(repositoryURL, branch, dockerfilePath)
	if err != nil {
		return "", err
	}
	tmpDirPath, err := os.MkdirTemp("", "filecache")
	if err != nil {
		return "", err
	}
	loadedPath := filepath.Join(tmpDirPath, "Dockerfile")
	err = os.WriteFile(loadedPath, data, 0755)
	if err != nil {
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
		&git.CloneOptions{URL: repositoryURL, NoCheckout: true},
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
		return []byte{}, err
	}

	fileHandle, err := fs.Open(path)
	if err != nil {
		return []byte{}, err
	}

	var data bytes.Buffer
	_, err = io.Copy(&data, fileHandle)
	if err != nil {
		return []byte{}, err
	}
	return data.Bytes(), nil
}
