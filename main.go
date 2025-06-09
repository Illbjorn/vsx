package main

import (
	"os"

	"github.com/illbjorn/argv"
	"github.com/illbjorn/echo"
	"github.com/illbjorn/vsx/gallery"
)

const (
	app = "vsx"
)

// TODO: Implement `config` subcommand to manually persist configuration values
// TODO: Implement signature verification of downloaded VSIX files (PKCS #1 / v1.5)
// TODO: Implement `update` subcommand
// TODO: Implement `backup` and `restore` subcommands
// TODO: Implement timeout support (init contexts, pass with timeout to CMD handlers)
// TODO: Implement `list` subcommand

func main() {
	// Parse command-line args
	//
	// Some of these values (ex: `extension-dir`) can be persistent configuration
	// items. Considering this, further down we'll merge these values into the
	// `Config` instance and persist them.
	cmd, err := argv.Parse(os.Args[1:])
	if err != nil {
		echo.Fatalf("Failed to parse input: %s.", err)
	}

	// Apply debug configuration if the [`--debug`, `-d`] flag was provided
	if _, ok := cmd.Flag(flagDebug, flagDebugShort); !ok {
		echo.SetFlags(echo.WithLevel, echo.WithColor)
	} else {
		echo.SetLevel(echo.LevelDebug)
		echo.SetFlags(
			echo.WithCallerFile,
			echo.WithCallerLine,
			echo.WithLevel,
			echo.WithColor,
		)
	}

	// Prepare configuration
	//
	// 1. Load the config file
	// 2. Load any environment configuration, clobber config file values with
	//    these
	// 3. Clobber any file/env configuration values with command-line arg values
	cfg, err := LoadConfigFile()
	if err != nil {
		echo.Info("No configuration found, proceeding with defaults.")
		cfg = new(Config)
	}
	cfg = LoadConfigEnv(cfg)
	cfg = MergeInputs(cfg, cmd)

	// Save the config now to reflect any command-line updates
	err = SaveConfigFile(cfg)
	if err != nil {
		echo.Errorf("Failed to save config: %s.", err)
	}

	// Init the Gallery client
	g := gallery.New(cfg.GalleryScheme, cfg.GalleryHost)

	// Exec the command
	err = Run(g, cfg, cmd)
	if err != nil {
		echo.Fatalf("ERROR: %s.", err)
	}
}
