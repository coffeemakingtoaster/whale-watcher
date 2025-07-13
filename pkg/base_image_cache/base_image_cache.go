package baseimagecache

import (
	"database/sql"
	"errors"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/base_image_cache/db"
)

type BaseImageCache struct {
	cacheDir string
	dbConn   *sql.DB
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
