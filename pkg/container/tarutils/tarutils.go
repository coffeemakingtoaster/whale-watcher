package tarutils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func GetAvailabeInTarData(data []byte) ([]string, error) {
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

func getBlobByPattern(reader *tar.Reader, searchValue string, headerNameTransformer func(string) string) ([]byte, error) {
	for true {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				return []byte{}, ValueNotFound{digest: searchValue, tarPath: "in place"}
			}
		}
		log.Debug().Msg(header.Name)
		if headerNameTransformer(header.Name) == searchValue {
			break
		}
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func GetBlobFromPathByDigest(path string, digest string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	return getBlobByPattern(tar.NewReader(f), digest, nameToBlobDigest)
}

func GetBlobFromDataByDigest(data []byte, digest string) ([]byte, error) {
	reader := bytes.NewReader(data)
	return getBlobByPattern(tar.NewReader(reader), digest, nameToBlobDigest)
}

func GetBlobFromDataByName(data []byte, name string) ([]byte, error) {
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
