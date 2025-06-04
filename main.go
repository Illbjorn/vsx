package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// Parse command line args
	args := ParseArgs()

	// Apply debug configuration if the [`--debug`, `-d`] flag was provided
	if args.Debug {
		echo.SetLevel(echo.LevelDebug)
		echo.SetFlags(
			echo.WithCallerFile,
			echo.WithCallerLine,
			echo.WithLevel,
			echo.WithColor,
		)
	}

	// Discern the command
	cmd := strings.ToLower(args.Positional[0])
	switch cmd {
	case cmdInstall: // No-op
		echo.Debugf("Received command [%s].", cmd)
	case cmdDownload: // No-op
		echo.Debugf("Received command [%s].", cmd)
	default:
		echo.Errorf("Received unknown command [%s].", cmd)
		echo.Error(Usage())
	}

	// Retrieve publisher and extension ID from the parsed positional args
	publisher, extensionID := parsePackage(args.Positional)

	// Apply default config
	//
	// If we didn't find a gallery scheme, use `https`
	if args.GalleryScheme == "" {
		args.GalleryScheme = "https"
	}
	// Version is required to fetch the extension, if one wasn't provided we can
	// use a default of `latest`
	if args.Version == "" {
		args.Version = "latest"
	}
	// If we didn't receive a path to save the vsix package to, default to
	// `[publisher]-[extensionID]-[version].vsix` in the current working directory
	if args.Output == "" {
		args.Output = fmt.Sprintf("%s-%s-%s.vsix", publisher, extensionID, args.Version)
	}
	// If we don't have an extension directory, try to locate one in the home
	// directory
	if cmd == cmdInstall && args.ExtensionDir == "" {
		// Get the home directory path
		home, err := os.UserHomeDir()
		assert(
			err == nil,
			"No extension directory was provided and we failed to locate the home "+
				"directory: %s.",
			err,
		)

		// Define the extension's subdirectory name (ex: `~/.vscode/extensions/[here]`)
		extDirName := fmt.Sprintf("%s-%s-%s", publisher, extensionID, args.Version)

		// Look for a `.vscode-oss` or `.vscode` directory in the user's home path
		extDirOSS := filepath.Join(home, ".vscode-oss")
		extDirMS := filepath.Join(home, ".vscode")
		if _, err := os.Lstat(extDirOSS); err == nil {
			args.ExtensionDir = filepath.Join(extDirOSS, "extensions", extDirName)
			echo.Debugf("Using extension directory [%s].", extDirOSS)

		} else if _, err := os.Lstat(extDirMS); err == nil {
			args.ExtensionDir = filepath.Join(extDirMS, "extensions", extDirName)
			echo.Debugf("Using extension directory [%s].", extDirMS)

		} else {
			echo.Debugf("Attempted: %s.", extDirOSS)
			echo.Debugf("Attempted: %s.", extDirMS)
			echo.Fatal("An extension directory was not provided and we failed to locate one.")
		}

		// Create the extension dir if needed
		err = os.MkdirAll(args.ExtensionDir, fileModeRWX)
		assert(
			err == nil || errors.Is(err, os.ErrExist),
			"Failed to create extension output directory [%s]: %s.",
			args.ExtensionDir, err,
		)
	}
	echo.Debugf("Using VS extension directory [%s].", args.ExtensionDir)

	// Init the Gallery
	g := NewGallery(args.GalleryScheme, args.GalleryHost)

	// Exec the command we received
	switch cmd {
	case "install":
		InstallExtension(g, args.ExtensionDir, publisher, extensionID, args.Version)

	case "download":
		DownloadExtension(g, args.Output, publisher, extensionID, args.Version)
	}
}

func parsePackage(args []string) (string, string) {
	assert(
		len(args) > 0,
		"Received no positional args, please provide the VS extension identifier (ex: `modular-mojotools.vscode-mojo`).",
	)

	// Get the extension identifier from the last arg
	publisherAndID := args[len(args)-1]
	echo.Debugf("Found raw package input [%s].", publisherAndID)

	// The extension identifier will be in the format `[publisher.extensionID]`,
	// we just slice on either side of the `.`
	i := strings.IndexByte(publisherAndID, '.')
	assert(i != -1,
		"Input [%s] is not a valid VS extension identifier (ex: `modular-mojotools.vscode-mojo`).",
		publisherAndID,
	)

	return publisherAndID[:i], publisherAndID[i+1:]
}
