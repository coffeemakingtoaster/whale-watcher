package baseimagecache

import (
	"database/sql"
	"errors"
	"path/filepath"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/base_image_cache/db"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
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

func NewBaseImageCache() *BaseImageCache {
	cfg := config.GetConfig()
	dbPath := filepath.Join(cfg.BaseImageCache.CacheLocation, "base_image_cache.db")
	conn, _ := db.LoadOrInitDB(dbPath)
	return &BaseImageCache{
		cacheDir: cfg.BaseImageCache.CacheLocation,
		dbConn:   conn,
	}
}
