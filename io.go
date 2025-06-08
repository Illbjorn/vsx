package main

import "os"

const (
	fileFlagsOverwrite = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	fileFlagsRead      = os.O_RDONLY
	fileModeRWX        = 0o700
	fileModeRW         = 0o600
)
