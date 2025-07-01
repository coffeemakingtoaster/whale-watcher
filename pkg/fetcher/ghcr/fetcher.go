package ghcr

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

var ghcrApiBase = "https://ghcr.io/v2"

type OCIDownloader struct {
	image       string
	tag         string
	destination string
	token       string
	namespace   string
}

func (od *OCIDownloader) ToString() string {
	return fmt.Sprintf("Image: %s Tag: %s Namespace: %s", od.image, od.tag, od.namespace)
}

func newOciDownloader(image, destination string) *OCIDownloader {
	components := strings.Split(image, ":")
	if len(components) < 2 {
		log.Error().Msg("Image tag needs to be provided")
		return nil
	}
	tag := components[len(components)-1]
	image = components[0]
	parts := strings.Split(image, "/")
	namespace := ""
	if len(parts) > 1 {
		namespace = strings.Join(parts[:len(parts)-1], "/")
	}
	return &OCIDownloader{
		image:       image,
		tag:         tag,
		destination: destination,
		namespace:   namespace,
	}
}

// For now this ONLY supports ghcr
func DownloadOciToPath(image, destination string) error {
	downloader := newOciDownloader(image, destination)
	if downloader == nil {
		return errors.New("Could not parse provided image, see logs for details ")
	}
	downloader.RefreshToken()
	/*
		if downloader.token == "" {
			return errors.New("Could not fetch auth token for repository, see logs for details")
		}
	*/
	manifest, err := downloader.GetManifest()
	if err != nil {
		return err
	}
	// default to first
	wanted := manifest.Manifests[0]
	for _, v := range manifest.Manifests {
		if v.Platform.OS == "linux" {
			wanted = v
		}
	}
	man, _ := downloader.GetSpecificManifest(wanted.Digest)
	writer := downloader.openTar()
	defer writer.Close()
	for i, v := range man.Layers {
		log.Debug().Int("Current layer", i).Int("total layers", len(man.Layers)).Msg("Pulling image")
		err := downloader.addLayerToTar(writer, v)
		if err != nil {
			log.Error().Err(err).Msg("Could not add layer to tar due to an error")
		}
	}
	manifestMetadata := tar.Header{
		Name:    fmt.Sprintf("blobs/%s", strings.Replace(wanted.Digest, ":", "/", 1)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	err = writeStructToTar(writer, &manifestMetadata, man)
	if err != nil {
		log.Error().Err(err).Msg("Could not write manifest to tar")
	}
	indexMetadata := tar.Header{
		Name:    "index.json",
		Mode:    0644,
		ModTime: time.Now(),
	}
	// only show manifest that was downloaded
	manifest.Manifests = []container.Manifest{wanted}
	err = writeStructToTar(writer, &indexMetadata, manifest)
	if err != nil {
		log.Error().Err(err).Msg("Could not write index to tar")
	}
	err = downloader.addConfigToTar(writer, man.Config.Digest)
	if err != nil {
		log.Error().Err(err).Msg("Could not write metadata to tar")
	}
	return nil
}

func (od *OCIDownloader) RefreshToken() {
	res, err := http.Get(
		fmt.Sprintf("https://ghcr.io/token?scope=repository:%s:pull", od.image),
	)

	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return
	}
	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Error().Err(err).Str("data", string(data)).Msg("Could not fetch auth token for container repository")
		return
	}
	val, ok := parsed["token"]
	if !ok {
		log.Error().Msgf("No token in repository response: %v", parsed)
		return
	}
	od.token = val.(string)
	log.Debug().Msg("OCI Downloader auth token refreshed")
}

func (od *OCIDownloader) getRequest(url string) (*http.Request, error) {
	if od.token == "" {
		od.RefreshToken()
		if od.token == "" {
			return nil, errors.New("Could not refresh token, check logs for details")
		}
		return od.getRequest(url)
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", od.token))
	return req, nil
}

func (od *OCIDownloader) GetManifest() (container.OCIImageIndex, error) {
	client := &http.Client{}
	req, err := od.getRequest(fmt.Sprintf("%s/%s/manifests/%s", ghcrApiBase, od.image, od.tag))
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json")
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch manifest for image")
		return container.OCIImageIndex{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch manifest for image")
		return container.OCIImageIndex{}, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return container.OCIImageIndex{}, err
	}
	var parsed container.OCIImageIndex
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return container.OCIImageIndex{}, err
	}
	log.Debug().Msg("OCI image manifest fetched")
	return parsed, nil
}

func (od *OCIDownloader) GetSpecificManifest(digest string) (container.OCIImageManifest, error) {
	client := &http.Client{}
	req, err := od.getRequest(fmt.Sprintf("%s/%s/manifests/%s", ghcrApiBase, od.image, digest))
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json")
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch manifest for image")
		return container.OCIImageManifest{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch manifest for image")
		return container.OCIImageManifest{}, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return container.OCIImageManifest{}, err
	}
	var parsed container.OCIImageManifest
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch auth token for container repository")
		return container.OCIImageManifest{}, err
	}
	log.Debug().Msg("OCI image manifest fetched")
	return parsed, nil

}

func (od *OCIDownloader) openTar() *tar.Writer {
	// TODO: Error handling
	file, err := os.Create(od.destination)
	if err != nil {
		log.Error().Err(err).Msg("Could not create tar file")
	}
	writer := tar.NewWriter(file)
	return writer
}

func (od *OCIDownloader) getLayerData(digest string) ([]byte, error) {
	client := &http.Client{}
	req, err := od.getRequest(fmt.Sprintf("%s/%s/blobs/%s", ghcrApiBase, od.image, digest))

	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Accept", "")
	res, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func writeToTar(writer *tar.Writer, header *tar.Header, data []byte) error {
	if err := writer.WriteHeader(header); err != nil {
		return err
	}

	buff := bytes.NewBuffer(data)

	_, err := io.Copy(writer, buff)
	if err != nil {
		return err
	}
	return nil
}

func (od *OCIDownloader) addLayerToTar(writer *tar.Writer, metadata container.LayerMetaData) error {
	data, err := od.getLayerData(metadata.Digest)
	if err != nil {
		return err
	}
	name := strings.Replace(metadata.Digest, ":", "/", 1)
	header := &tar.Header{
		Name:    fmt.Sprintf("blobs/%s", name),
		Mode:    0644,
		Size:    int64(metadata.Size),
		ModTime: time.Now(),
	}
	return writeToTar(writer, header, data)
}

func (od *OCIDownloader) addConfigToTar(writer *tar.Writer, digest string) error {
	client := &http.Client{}
	req, err := od.getRequest(fmt.Sprintf("%s/%s/blobs/%s", ghcrApiBase, od.image, digest))

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "")
	res, err := client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var cfg container.ImageMetadata

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	header := tar.Header{
		Name:    fmt.Sprintf("blobs/%s", strings.Replace(digest, ":", "/", 1)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	return writeStructToTar(writer, &header, cfg)
}

func writeStructToTar[T any](writer *tar.Writer, header *tar.Header, data T) error {
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}
	header.Size = int64(len(content))
	return writeToTar(writer, header, content)
}
