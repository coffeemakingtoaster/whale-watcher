package db

const initStatement = `
	CREATE TABLE IF NOT EXISTS image_package_lookup(image TEXT NOT NULL, package TEXT NOT NULL, version TEXT, base TEXT);
	CREATE TABLE IF NOT EXISTS image_digest_lookup(image TEXT NOT NULL, digest TEXT NOT NULL);
`
const insertImagePackageStatement = "INSERT INTO image_package_lookup(image, package, version, base) VALUES ('%s', '%s', '%s', '%s')"
const insertDigestStatement = "INSERT INTO image_digest_lookup(image, digest) VALUES ('%s','%s')"
