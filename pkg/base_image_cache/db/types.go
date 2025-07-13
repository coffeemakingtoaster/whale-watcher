package db

type BaseImagePackageEntry struct {
	Image          string
	Base           string
	Package        string
	PackageVersion string
}

type HitInfo struct {
	Image           string
	MatchedPackages int
}
