package tarutils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type ValueNotFound struct {
	digest  string
	tarPath string
}

func (e ValueNotFound) Error() string {
	return fmt.Sprintf("Digest %s not found in tarfile %s", e.digest, e.tarPath)
}

func ParseJsonBytesIntoInterface[T any](data []byte) (T, error) {
	var parsed T
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return *new(T), err
	}
	return parsed, nil
}

func ParseJsonReaderIntoInterface[T any](reader *tar.Reader) (T, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return *new(T), err
	}
	return ParseJsonBytesIntoInterface[T](data)
}

func GetAvailabeInTar(data []byte) ([]string, error) {
	reader := tar.NewReader(bytes.NewReader(data))
	available := []string{}
	for true {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return available, err
		}
		available = append(available, header.Name)
	}
	return available, nil
}

func getBlobByPattern(reader *tar.Reader, searchValue string, headerNameTransformer func(string) string) (*tar.Header, *tar.Reader, error) {
	for true {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		log.Debug().Str("name", header.Name).Str("format", header.Format.String()).Int("type", int(header.Typeflag)).Send()
		if headerNameTransformer(header.Name) == searchValue {
			return header, reader, nil
		}
	}
	return nil, nil, ValueNotFound{digest: searchValue, tarPath: "in place"}
}

func GetBlobFromFileByName(tarpath, searchValue string) (*tar.Header, *tar.Reader, error) {
	f, err := os.Open(tarpath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	reader := tar.NewReader(f)

	return getBlobByPattern(reader, searchValue, func(s string) string { return s })
}

func GetBlobFromFileByDigest(tarpath, digest string) (*tar.Header, *tar.Reader, error) {
	f, err := os.Open(tarpath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	reader := tar.NewReader(f)
	return getBlobByPattern(reader, digest, nameToBlobDigest)
}

func GetBlobFromDataByDigest(data []byte, digest string) (*tar.Header, *tar.Reader, error) {
	reader := bytes.NewReader(data)
	return getBlobByPattern(tar.NewReader(reader), digest, nameToBlobDigest)
}

func GetBlobFromDataByName(data []byte, name string) (*tar.Header, *tar.Reader, error) {
	reader := bytes.NewReader(data)
	return getBlobByPattern(tar.NewReader(reader), name, func(s string) string { return s })
}

func nameToBlobDigest(path string) string {
	digest := strings.TrimPrefix(path, "blobs/")
	return strings.ReplaceAll(digest, "/", ":")
}

func UngzipBlob(raw []byte) ([]byte, error) {
	reader := bytes.NewReader(raw)
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ungzip data")
		return []byte{}, err
	}
	data, err := io.ReadAll(gzreader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read ungzip data")
		return []byte{}, err
	}
	return data, nil
}
