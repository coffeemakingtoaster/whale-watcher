package container

import (
	"fmt"
	"strings"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/layerfs"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/tarutils"
)

type Layer struct {
	Digest     string
	Command    string
	tarPath    string
	FileSystem layerfs.LayerFS
}

func (l *Layer) ToString() string {
	return fmt.Sprintf("[%s](%s) %s", l.Digest, l.tarPath, l.FileSystem.ToString())
}

// Use the command to estimate installed packages
// This just looks for known package manager install paths...i.e. this may not be reliable
// Supported package managers: apt/apt-get, brew, apk
func (l *Layer) GetInstalledPackagesEstimate() []string {
	cmds := strings.SplitSeq(l.Command, "&&")
	packages := []string{}
	for cmd := range cmds {
		cmd = strings.TrimPrefix(cmd, " ")
		cmd = strings.TrimSuffix(cmd, " ")
		segments := strings.SplitSeq(cmd, ";")
		for segment := range segments {
			segment = strings.TrimSpace(segment)
			// apt
			if strings.HasPrefix(segment, "apt") || strings.HasPrefix(segment, "apt-get") || strings.HasPrefix(segment, "brew") || strings.HasPrefix(segment, "apk") {
				segment = strings.TrimPrefix(segment, "apt-get")
				segment = strings.TrimPrefix(segment, "apt")
				segment = strings.TrimPrefix(segment, "apk")
				segment = strings.TrimPrefix(segment, "brew")
				segment = strings.TrimPrefix(segment, "install")
				segment = strings.TrimSpace(segment)
				pkgs := strings.Split(segment, " ")
				for i := range pkgs {
					pkgs[i] = strings.TrimSpace(pkgs[i])
				}
				packages = append(packages, pkgs...)
			}
		}
	}
	return cleanFromParamsAndFlags(packages)
}

func cleanFromParamsAndFlags(input []string) []string {
	res := []string{}
	for _, v := range input {
		if !strings.HasPrefix(v, "-") {
			res = append(res, v)
		}
	}
	return res
}

func NewLayer(loadedTar *tarutils.LoadedTar, digest, command string, isGzip bool) *Layer {
	return &Layer{
		Command:    command,
		Digest:     digest,
		tarPath:    loadedTar.TarPath,
		FileSystem: layerfs.NewLayerFS(loadedTar, digest, isGzip),
	}
}
