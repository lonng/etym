package fs

import "os"

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func EnsureDir(path string) {
	if IsExists(path) {
		return
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}
}
