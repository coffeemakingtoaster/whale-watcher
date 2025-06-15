package layerfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/armon/go-radix"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container/tarutils"
)

// This fs loads the files on demand and does NOT cache the entire blob
// i.e. every open request (that is not cached) leads to a whole read of the tar again...
type LayerFS struct {
	tarPath              string
	digest               string
	fileCache            map[string]*LayerFS
	maxFileCacheCapacity int
	isGzip               bool
	lookupRadix          radix.Tree
}

// LayerFSFile wraps a byte slice and implements the fs.File interface.
type LayerFSFile struct {
	data   []byte
	offset int64
	name   string
}

// Read reads up to len(p) bytes into p.
func (f *LayerFSFile) Read(p []byte) (int, error) {
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// Close is a no-op for CustomFile.
func (f *LayerFSFile) Close() error {
	return nil
}

// Stat returns a dummy fs.FileInfo.
func (f *LayerFSFile) Stat() (fs.FileInfo, error) {
	return &LayerFSFileInfo{size: int64(len(f.data)), name: f.name}, nil
}

// LayerFSFileInfo implements fs.FileInfo for CustomFile.
type LayerFSFileInfo struct {
	size int64
	name string
}

func (fi *LayerFSFileInfo) Name() string       { return fi.name }
func (fi *LayerFSFileInfo) Size() int64        { return fi.size }
func (fi *LayerFSFileInfo) Mode() fs.FileMode  { return 0444 }
func (fi *LayerFSFileInfo) ModTime() time.Time { return time.Now() }
func (fi *LayerFSFileInfo) IsDir() bool        { return false }
func (fi *LayerFSFileInfo) Sys() any           { return nil }

func (lfs LayerFS) Open(name string) (fs.File, error) {
	if _, ok := lfs.lookupRadix.Get(name); !ok {
		return nil, errors.New("File not found")
	}
	data, err := tarutils.GetBlobFromFileByDigest(lfs.tarPath, lfs.digest)
	if err != nil {
		return nil, err
	}
	if lfs.isGzip {
		data, err = tarutils.UngzipBlob(data)
		if err != nil {
			return nil, err
		}
	}
	fileData, err := tarutils.GetBlobFromDataByDigest(data, name)
	if err != nil {
		return nil, err
	}
	return &LayerFSFile{data: fileData, name: name}, nil
}

// Only supports absolute paths
func (lfs *LayerFS) Ls(path string) []string {
	result := []string{}
	walkFn := func(s string, _ interface{}) bool {
		result = append(result, s)
		return false
	}
	lfs.lookupRadix.WalkPrefix(path, walkFn)
	return result
}

func getAllFiles(tarpath, digest string, isGzip bool) ([]string, error) {
	data, err := tarutils.GetBlobFromFileByDigest(tarpath, digest)
	if err != nil {
		return nil, err
	}
	if isGzip {
		data, err = tarutils.UngzipBlob(data)
		if err != nil {
			return nil, err
		}
	}
	return tarutils.GetAvailabeInTar(data)
}

func (lfs *LayerFS) ToString() string {
	return fmt.Sprintf("FS with %d files!", lfs.lookupRadix.Len())
}

func NewLayerFS(tarPath, digest string, isGzip bool) LayerFS {
	availableFiles, _ := getAllFiles(tarPath, digest, isGzip)
	m := make(map[string]any)
	for _, v := range availableFiles {
		m[v] = 1
	}
	return LayerFS{
		tarPath:              tarPath,
		digest:               digest,
		fileCache:            make(map[string]*LayerFS),
		maxFileCacheCapacity: 5,
		isGzip:               isGzip,
		lookupRadix:          *radix.NewFromMap(m),
	}
}
