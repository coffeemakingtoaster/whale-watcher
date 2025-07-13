package baseimagecache

import (
	"database/sql"
	"errors"

	"github.com/coffeemakingtoaster/oci-pull-go/pkg/pull"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/base_image_cache/db"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner"
)

type BaseImageCache struct {
	cacheDir string
	dbConn   *sql.DB
}

func (bic *BaseImageCache) IngestImage(image string) error {
	log.Debug().Str("image", image).Msg("Inserting image into base image cache")
	// Download OCI
	pwd := runner.GetReferencingWorkingDirectoryInstance()
	defer pwd.Free()
	destination := pwd.GetAbsolutePath("image.tar")
	err := pull.PullToPath(image, destination)
	if err != nil {
		log.Error().Err(err).Msgf("Could not download image %s", image)
		return err
	}
	// Build list of installed packages based on OCI
	ociImage, err := container.ContainerImageFromOCITar(destination)
	if err != nil {
		return err
	}
	packages := map[string]bool{}
	digests := make([]string, len(ociImage.Layers))
	log.Debug().Msg("Building package and digest list")
	for i := range len(ociImage.Layers) - 1 {
		layer := ociImage.Layers[i+1]
		digests[i] = layer.Digest
		for _, pkg := range layer.GetInstalledPackagesEstimate() {
			if _, ok := packages[pkg]; !ok {
				packages[pkg] = true
			}
		}
	}

	// Insert into db
	for k := range packages {
		err := db.AddImagePackage(bic.dbConn, image, k, "")
		if err != nil {
			log.Error().Err(err).Msg("Could not add pkg to db")
		}
	}
	log.Debug().Int("package_count", len(packages)).Msg("Packages inserted")
	for _, digest := range digests {
		err := db.AddImageDigest(bic.dbConn, image, digest)
		if err != nil {
			log.Error().Err(err).Msg("Could not add pkg to db")
		}
	}
	log.Debug().Int("digest_count", len(digests)).Msg("Digests inserted")
	return nil
}

func (bic *BaseImageCache) QueryImage(image string) (db.BaseImagePackageEntry, error) {
	return db.QueryElemByProperties(bic.dbConn, &db.BaseImagePackageEntry{Image: image})
}

func (bic *BaseImageCache) GetImageByDigest(digest string) (string, error) {
	images, err := db.QueryImageByDigest(bic.dbConn, digest)
	if err != nil {
		return "", err
	}
	if len(images) == 0 {
		return "", errors.New("No image found with that digest")
	}
	return images[0], nil
}

func (bic *BaseImageCache) GetClosestDependencyImage(packages []string) (string, error) {
	hits, err := db.GetSortedByPackages(bic.dbConn, packages, []string{})
	if err != nil {
		return "", err
	}
	if len(hits) == 0 {
		return "", errors.New("No image found")
	}
	return hits[0].Image, nil
}
func (bic *BaseImageCache) GetClosestDependencyImageWithBase(base string, packages []string) error {
	return nil
}

func NewBaseImageCache(cachePath string) *BaseImageCache {
	conn, _ := db.LoadOrInitDB(cachePath)
	return &BaseImageCache{
		cacheDir: cachePath,
		dbConn:   conn,
	}
}
