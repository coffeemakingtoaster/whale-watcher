package tarutils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

type ValueNotFound struct {
	digest  string
	tarPath string
}

type LoadedTar struct {
	TarPath string
	data    map[string][]byte
	keys    []string
	loaded  bool
}

func LoadTar(tarPath string) *LoadedTar {
	return &LoadedTar{
		TarPath: tarPath,
		data:    make(map[string][]byte),
		keys:    make([]string, 0),
	}
}

func (lt *LoadedTar) load() error {
	if lt.loaded {
		return nil
	}

	f, err := os.Open(lt.TarPath)
	if err != nil {
		return err
	}

	defer f.Close()
	reader := tar.NewReader(f)

	for true {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		data, err := io.ReadAll(reader)

		if err != nil {
			return err
		}

		lt.data[header.Name] = data
		lt.keys = append(lt.keys, header.Name)
	}
	lt.loaded = true
	return nil
}

func (lt *LoadedTar) Unload() {
	lt.data = map[string][]byte{}
	lt.keys = []string{}
}

func (lt *LoadedTar) GetAvailabe() ([]string, error) {
	err := lt.load()
	return lt.keys, err
}

func (e ValueNotFound) Error() string {
	return fmt.Sprintf("Digest %s not found in tarfile %s", e.digest, e.tarPath)
}

func (lt *LoadedTar) getBlobByPattern(searchValue string, headerNameTransformer func(string) string) ([]byte, error) {
	err := lt.load()
	if err != nil {
		return []byte{}, err
	}
	for _, k := range lt.keys {
		if headerNameTransformer(k) == searchValue {
			return lt.data[k], nil
		}
	}
	return []byte{}, ValueNotFound{digest: searchValue, tarPath: "in place"}
}

func (lt *LoadedTar) GetBlobFromFileByName(searchValue string) ([]byte, error) {
	return lt.getBlobByPattern(searchValue, func(s string) string { return s })
}

func (lt *LoadedTar) GetBlobFromFileByDigest(digest string) ([]byte, error) {
	return lt.getBlobByPattern(digest, nameToBlobDigest)
}
