package targets

import "strings"

const (
	COMMAND_UTIL_LEVEL = iota
	FS_UTIL_LEVEL
	OS_UTIL_LEVEL
)

const COMMAND_UTIL_VALUE = "command"
const FS_UTIL_VALUE = "fs"
const OS_UTIL_VALUE = "os"

func GetTargetLevel(target string) int {
	switch strings.ToLower(target) {
	case COMMAND_UTIL_VALUE:
		return COMMAND_UTIL_LEVEL
	case FS_UTIL_VALUE:
		return FS_UTIL_LEVEL
	default:
		return OS_UTIL_LEVEL
	}
}
