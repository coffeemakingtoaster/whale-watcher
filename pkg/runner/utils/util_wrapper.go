package utils

import (
	"embed"
)

//go:embed _fs_util_build/*
var Fsutil embed.FS

//go:embed _os_util_build/*
var Osutil embed.FS

//go:embed _command_util_build/*
var Cmdutil embed.FS

//go:embed _fix_util_build/*
var Fixutil embed.FS
