package baseimagecache

import (
	"database/sql"
	"errors"

	"github.com/coffeemakingtoaster/oci-pull-go/pkg/pull"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/base_image_cache/db"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner"
	fsutil "iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner/fs_util"
)

type BaseImageCache struct {
	cacheDir string
	dbConn   *sql.DB
}

func (bic *BaseImageCache) IngestImage(image string) error {
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

	// Insert into db
	return nil
}
func (bic *BaseImageCache) QueryImage(image string) (db.BaseImagePackageEntry, error) {
	return db.QueryElemByProperties(bic.dbConn, &db.BaseImagePackageEntry{Image: image})
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
