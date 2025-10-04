package fetcher

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func FetchContainerFiles() (string, string, string) {
	var dockerfilePath string
	var ociPath string
	var dockerPath string
	var err error

	if viper.GetString("target.repository") == "" {
		dockerfilePath = viper.GetString("target.dockerfile")
	} else {
		dockerfilePath, err = loadDockerfileFromRepository(viper.GetString("target.repository"), viper.GetString("target.branch"), viper.GetString("target.dockerfile"))

		if err != nil {
			log.Warn().Err(err).Msg("Could not load dockerfile from repository")
		}
	}

	if !config.AllowsTarget("os") && !config.AllowsTarget("fs") {
		log.Info().Msg("Fs and Os targets disallowed, skipping image download & load")
	} else {
		if viper.GetString("target.image") == "" {
			log.Debug().Msg("Using local files for tar paths")
			ociPath = viper.GetString("target.ocipath")
			dockerPath = viper.GetString("target.dockerpath")
		} else {
			ociPath, dockerPath, err = loadImageFromRegistry(viper.GetString("target.image"), viper.GetBool("target.insecure"))
			if err != nil {
				log.Warn().Err(err).Msg("Could not load image from repository")
			}
		}
	}

	return dockerfilePath, ociPath, dockerPath
}

func loadImageFromRegistry(image string, insecure bool) (string, string, error) {
	log.Info().Str("image", image).Msg("Downloading image from registry")
	tmpDirPath, err := os.MkdirTemp("", "filecache")
	if err != nil {
		if !os.IsExist(err) {
			return "", "", err
		}
	}

	destinationOci := filepath.Join(tmpDirPath, "image.tar")
	err = LoadTarToPath(image, destinationOci, "oci", insecure)
	if err != nil {
		return "", "", err
	}

	destinationDocker := filepath.Join(tmpDirPath, "image_docker.tar")
	err = LoadTarToPath(image, destinationDocker, "docker", insecure)
	if err != nil {
		return "", "", err
	}
	log.Info().Str("image", image).Str("dockerTarPath", destinationDocker).Str("ociTarPath", destinationOci).Msg("Successful download")
	return destinationOci, destinationDocker, nil
}

func LoadTarToPath(image, destination, format string, insecure bool) error {
	format = strings.ToLower(format)
	if format != "oci" && format != "docker" {
		return fmt.Errorf("unsupported format: %s (supported: 'oci', 'docker')", format)
	}

	var ref name.Reference
	var err error
	var remoteOptions []remote.Option

	if !insecure {
		ref, err = name.ParseReference(image)
	} else {
		log.Warn().Msg("Insecure was enabled in config, performing insecure image pull (http)")
		ref, err = name.ParseReference(image, name.Insecure)
		if err != nil {
			return err
		}
		//remoteOptions = append(remoteOptions, remote.WithAuth(authn.Anonymous))
	}

	if err != nil {
		return err
	}

	img, err := remote.Image(ref, remoteOptions...)
	if err != nil {
		return err
	}
	switch format {
	case "docker":
		log.Info().Str("image", image).Msg("Saving docker tarball")
		return saveAsDockerTarball(ref, img, destination)
	case "oci":
		log.Info().Str("image", image).Msg("Saving oci tarball")
		return saveAsOCITarball(img, destination)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// saveAsDockerTarball saves the image as a Docker-compatible tarball
func saveAsDockerTarball(ref name.Reference, img v1.Image, destination string) error {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tarball.Write(ref, img, file)
	if err != nil {
		log.Error().Err(err).Msgf("Could not create Docker tarball for image %s", ref.String())
		return err
	}

	return nil
}

// saveAsOCITarball saves the image as an OCI-compliant tarball
func saveAsOCITarball(img v1.Image, destination string) error {
	// Create a temporary directory for the OCI layout
	tempDir, err := os.MkdirTemp("", "oci-layout-*")
	if err != nil {
		return err
	}

	// Ensure cleanup of temporary directory
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			log.Warn().Err(removeErr).Msg("Failed to remove temporary directory")
		}
	}()

	// Create OCI layout in temp directory
	layoutPath, err := layout.Write(tempDir, empty.Index)
	if err != nil {
		return err
	}

	// Append the image to the layout
	err = layoutPath.AppendImage(img)
	if err != nil {
		log.Error().Err(err).Msg("Could not append image to OCI layout")
		return err
	}

	// Create the destination tar file
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create tar writer (uncompressed)
	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	// Walk through the OCI layout directory and add files to tar
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from temp directory
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		// If it's a regular file, write its content
		if info.Mode().IsRegular() {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			_, err = io.Copy(tarWriter, srcFile)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Could not create OCI-compliant tar")
		return err
	}

	return nil
}

func loadDockerfileFromRepository(repositoryURL, branch, dockerfilePath string) (string, error) {
	data, err := GetFileFromRepository(repositoryURL, branch, dockerfilePath)
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

func GetFileFromRepository(repositoryURL, branch, path string) ([]byte, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	log.Debug().Str("url", repositoryURL).Msg("Cloning Repository")

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

	log.Debug().Str("path", path).Msg("Getting file")

	err = w.Checkout(&git.CheckoutOptions{SparseCheckoutDirectories: []string{dockerfileDirectory}, Branch: plumbing.ReferenceName(branchReference)})
	if err != nil {
		log.Error().Msg("Could not checkout")
		return []byte{}, err
	}

	log.Debug().Str("path", path).Msg("Reading file")

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
