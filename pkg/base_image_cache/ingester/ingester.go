package ingester

import (
	"path/filepath"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/base_image_cache/db"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/container"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/fetcher"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	"github.com/rs/zerolog/log"
)

func IngestImage(image string) error {
	log.Info().Str("image", image).Msg("Inserting image into base image cache")
	cfg := config.GetConfig()
	dbConn, err := db.LoadOrInitDB(filepath.Join(cfg.BaseImageCache.CacheLocation, "base_image_cache.db"))
	if err != nil {
		return err
	}
	elem, err := db.QueryElemByProperties(dbConn, &db.BaseImagePackageEntry{Image: image})
	if len(elem.Package) > 0 {
		log.Info().Str("image", image).Msg("Already present")
		return nil
	}
	// Download OCI
	pwd := runner.GetReferencingWorkingDirectoryInstance()
	defer pwd.Free()
	destination := pwd.GetAbsolutePath("image.tar")
	err = fetcher.LoadTarToPath(image, destination, "oci", false)
	if err != nil {
		log.Error().Err(err).Msgf("Could not download image %s", image)
		return err
	}

	packages, digests, err := getPackagesAndDigests(destination)
	if err != nil {
		return err
	}

	// Insert into db
	for k := range packages {
		err := db.AddImagePackage(dbConn, image, k, "")
		if err != nil {
			log.Error().Err(err).Msg("Could not add pkg to db")
		}
	}
	log.Debug().Int("package_count", len(packages)).Msg("Packages inserted")
	for _, digest := range digests {
		err := db.AddImageDigest(dbConn, image, digest)
		if err != nil {
			log.Error().Err(err).Msg("Could not add pkg to db")
		}
	}
	log.Debug().Int("digest_count", len(digests)).Msg("Digests inserted")
	return nil
}

func getPackagesAndDigests(destination string) (map[string]bool, []string, error) {
	// Build list of installed packages based on OCI
	ociImage, err := container.ContainerImageFromOCITar(destination)
	if err != nil {
		return make(map[string]bool), make([]string, 0), err
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
	return packages, digests, nil
}
