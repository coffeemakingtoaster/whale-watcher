package util

import "os"

var UnsafeModeEnvVar = "WHALE_WATCHER_ALLOW_UNSAFE"

func IsUnsafeMode() bool {
	return os.Getenv(UnsafeModeEnvVar) == "true"
}
