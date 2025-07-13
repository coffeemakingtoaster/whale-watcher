package db

const initStatement = "CREATE TABLE IF NOT EXIST image_package_lookup(image TEXT NOT NULL, package TEXT NOT NULL, version TEXT, base TEXT)"
const insertImagePackageStatement = "INSERT INTO image_package_lookup(image, package, version, base) VALUES ('%s', '%s', '%s', '%s')"
