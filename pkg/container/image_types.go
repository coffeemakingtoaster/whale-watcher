package container

import "time"

// OCIImageIndex represents the structure of the OCI image index JSON.
type OCIImageIndex struct {
	SchemaVersion int        `json:"schemaVersion"`
	MediaType     string     `json:"mediaType"`
	Manifests     []Manifest `json:"manifests"`
}

// Manifest represents the structure of each manifest in the OCI image index.
type Manifest struct {
	MediaType   string      `json:"mediaType"`
	Digest      string      `json:"digest"`
	Size        int         `json:"size"`
	Annotations Annotations `json:"annotations"`
	Platform    Platform    `json:"platform"`
}

// Annotations represents the annotations in the manifest.
type Annotations struct {
	Created time.Time `json:"org.opencontainers.image.created"`
}

// Platform represents the platform details in the manifest.
type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

// OCIImageManifest represents the structure of the OCI image manifest map.
type OCIImageManifest struct {
	Config        Config          `json:"config"`
	Layers        []LayerMetaData `json:"layers"`
	MediaType     string          `json:"mediaType"`
	SchemaVersion int             `json:"schemaVersion"`
}

// Config represents the configuration details in the OCI image manifest.
type Config struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
}

// Layer represents each layer in the OCI image manifest.
type LayerMetaData struct {
	Digest    string  `json:"digest"`
	MediaType string  `json:"mediaType"`
	Size      float64 `json:"size"`
}

type ImageMetadata struct {
	Architecture string `json:"architecture"`
	Config       struct {
		Env        []string `json:"Env"`
		Entrypoint []string `json:"Entrypoint"`
		WorkingDir string   `json:"WorkingDir"`
	} `json:"config"`
	Created time.Time `json:"created"`
	History []struct {
		Created    time.Time `json:"created"`
		CreatedBy  string    `json:"created_by"`
		Comment    string    `json:"comment"`
		EmptyLayer bool      `json:"empty_layer,omitempty"`
	} `json:"history"`
	Os     string `json:"os"`
	Rootfs struct {
		Type    string   `json:"type"`
		DiffIds []string `json:"diff_ids"`
	} `json:"rootfs"`
}
