package fsutil

import "os"

func Dir_content_count(dirPath string) int {
	res, err := os.ReadDir(dirPath)
	if err != nil {
		return -1
	}
	return len(res)
}

func Name() string {
	return "fs_util"
}

func main() {}
