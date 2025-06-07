package main

import (
	"os"

	"github.com/illbjorn/echo"
)

const (
	app = "vsx"
)

// TODO:
//
// - Implement `argv` and `gallery` packages
// 	 - Retire `main` package `tokenizeCMD`, `parseCMD` and `NewGallery`
//     functionality
// - Consolidate `argDefaults` into a more agnostic interface
// - Implement timeout support (init contexts, pass with timeout to CMD
//   handlers)
// - Implement signature verification of downloaded VSIX files (PKCS #1 / v1.5)
// - Implement `backup` subcommand
// - Implement `update` subcommand

func main() {
	// Attempt to load the config file from disk
	cfg, err := LoadConfigFile()
	if err != nil {
		echo.Info("No configuration found, proceeding with defaults.")
		cfg = new(Config)
	}

	// We clobber any on-disk values with values found in the environment
	cfg = LoadConfigEnv(cfg)

	// Save the config when we're done
	defer SaveConfigFile(cfg)

	// Parse command-line args
	//
	// Both execution-specific and broader (persistent) VSX configuration can be
	// provided via command-line args. So we merge the values which could be
	// persistent into the config (clobbering).
	cmd, flags, args := parseCMD(os.Args[1:])

	// Clobber config values with any found in args.
	cfg = mergeInputs(cfg, flags)

	// Apply debug configuration if the [`--debug`, `-d`] flag was provided
	if _, ok := flags[flagDebug]; ok {
		echo.SetLevel(echo.LevelDebug)
		echo.SetFlags(
			echo.WithCallerFile,
			echo.WithCallerLine,
			echo.WithLevel,
			echo.WithColor,
		)
	}

	// Init the Gallery client
	g := NewGallery(cfg.GalleryScheme, cfg.GalleryHost)

	if cmd != "" {
		// If we received a command, run it and exit
		//
		// Exec the command
		err := RunCMD(g, cfg, cmd, flags, args)
		if err != nil {
			echo.Fatalf("Failed [%s]: %w.", cmd, err)
		}
	} else {
		// Otherwise, enter the REPL
		EnterREPL(g, cfg)
	}
}
