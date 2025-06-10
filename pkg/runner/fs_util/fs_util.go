package fsutil

import "os"

type FsUtils struct {
	// TODO: Implement me
	Fs string
}

// Setup function used for instantiating util struct
func Setup() FsUtils {
	return FsUtils{}
}

func (ou FsUtils) Dir_content_count(dirPath string) int {
	res, err := os.ReadDir(dirPath)
	if err != nil {
		return -1
	}
	return len(res)
}

func (ou FsUtils) Name() string {
	return "fs_util"
}

func main() {}
