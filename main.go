package main

import (
	"github.com/illbjorn/echo"
)

const (
	cmdInstall  = "install"
	cmdDownload = "download"
)

// TODO:
//
// - Implement timeout support (init contexts, pass with timeout to CMD
//   handlers)
// - Implement signature verification of downloaded VSIX files (PKCS #1 / v1.5)
// - Implement `backup` subcommand
// - Implement `update` subcommand

func main() {
	// Load the config
	cfg := LoadConfig()

	// Apply debug configuration if the [`--debug`, `-d`] flag was provided
	if cfg.Debug {
		echo.SetLevel(echo.LevelDebug)
		echo.SetFlags(
			echo.WithCallerFile,
			echo.WithCallerLine,
			echo.WithLevel,
			echo.WithColor,
		)
	}

	// Exec the command
	Exec(cfg)
}
