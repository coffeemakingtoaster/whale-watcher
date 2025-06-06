package main

import "C"
import "os"

func dir_content_count(dirPath string) int {
	res, err := os.ReadDir(dirPath)
	if err != nil {
		return -1
	}
	return len(res)
}

func main() {}
