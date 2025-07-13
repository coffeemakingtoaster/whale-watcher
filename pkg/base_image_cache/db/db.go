package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func initDb(dbConn *sql.DB) error {
	_, err := dbConn.Exec(initStatement)
	return err
}

func LoadOrInitDB(cacheDir string) (*sql.DB, error) {
	dbPath := filepath.Join(cacheDir, "base_image_cache.db")
	needsInit := false
	if _, err := os.Open(dbPath); os.IsNotExist(err) {
		log.Info().Msg("Cache did not exist, creating cache...")
		needsInit = true
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error().Err(err).Msg("Could not open db connection")
	}

	if needsInit {
		err := initDb(db)
		if err != nil {
			log.Error().Err(err).Msg("Could not init DB")
		}
	}
	return db, nil
}

func ExecStatement(conn *sql.DB, statement string) error {
	tx, err := conn.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(statement)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func QueryElemByProperties(conn *sql.DB, partial *BaseImagePackageEntry) (BaseImagePackageEntry, error) {
	properties := []string{}
	values := []string{}
	if len(partial.Base) > 0 {
		properties = append(properties, "base")
		values = append(values, partial.Base)
	}
	if len(partial.Image) > 0 {
		properties = append(properties, "image")
		values = append(values, partial.Image)
	}
	if len(partial.Package) > 0 {
		properties = append(properties, "package")
		values = append(values, partial.Package)
	}
	if len(partial.PackageVersion) > 0 {
		properties = append(properties, "package_version")
		values = append(values, partial.PackageVersion)
	}
	query := "SELECT * FROM image_package_lookup WHERE"
	for i := range properties {
		query += fmt.Sprintf("%s = '%s'", properties[i], values[i])
	}
	res, err := DoQuery(conn, query)
	if err != nil {
		return BaseImagePackageEntry{}, err
	}
	if len(res) == 0 {
		return BaseImagePackageEntry{}, errors.New("No row found")
	}
	return res[0], nil
}

func GetSortedByPackages(conn *sql.DB, packages []string, versions []string) ([]HitInfo, error) {
	if len(versions) > 0 && len(versions) != len(packages) {
		return []HitInfo{}, errors.New("Amount of versions must either match amount of packages or be 0")
	}
	query := "SELECT * FROM image_package_lookup WHERE"
	if len(versions) > 0 {
		log.Warn().Msg("Not implemented!")
	} else {
		query += fmt.Sprintf(" package IN (%+q)", packages)
	}
	entries, err := DoQuery(conn, query)
	if err != nil {
		return []HitInfo{}, nil
	}
	counter := map[string]int{}
	for _, entry := range entries {
		if _, ok := counter[entry.Image]; !ok {
			counter[entry.Image] = 0
		}
		counter[entry.Image]++
	}
	result := make([]HitInfo, len(counter))
	index := 0
	for k, v := range counter {
		result[index] = HitInfo{
			Image:           k,
			MatchedPackages: v,
		}
	}
	sort.Slice(result, func(a, b int) bool {
		return result[a].MatchedPackages < result[b].MatchedPackages
	})
	return result, nil
}

func AddImagePackage(conn *sql.DB, image, pkg, version string) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(insertImagePackageStatement, image, pkg, version, "")
	_, err = tx.Exec(stmt)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func AddImageDigest(conn *sql.DB, image, digest string) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(insertDigestStatement, image, digest)
	_, err = tx.Exec(stmt)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func DoQuery(conn *sql.DB, query string) ([]BaseImagePackageEntry, error) {
	rows, err := conn.Query(query)
	if err != nil {
		return []BaseImagePackageEntry{}, err
	}
	defer rows.Close()
	var res []BaseImagePackageEntry
	for rows.Next() {
		var baseImagePackgeEntry BaseImagePackageEntry
		if err := rows.Scan(&baseImagePackgeEntry.Image, &baseImagePackgeEntry.Package, &baseImagePackgeEntry.PackageVersion, &baseImagePackgeEntry.Base); err != nil {
			return res, err
		}
		res = append(res, baseImagePackgeEntry)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

func QueryImageByDigest(conn *sql.DB, digest string) ([]string, error) {
	rows, err := conn.Query(fmt.Sprintf("SELECT image FROM image_digest_lookup WHERE digest = '%s'", digest))
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var image string
		if err := rows.Scan(&image); err != nil {
			return res, err
		}
		res = append(res, image)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil

}
