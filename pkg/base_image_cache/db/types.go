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
	TotalPackages   int
}

func (hi *HitInfo) CalcScore() int {
	return hi.MatchedPackages * (hi.MatchedPackages / hi.TotalPackages)
}
