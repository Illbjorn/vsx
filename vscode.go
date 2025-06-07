package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illbjorn/echo"
)

func ExtensionDir(publisher, extensionID, version string) string {
	// Get the home directory path
	home, err := os.UserHomeDir()
	must(
		err == nil,
		"No extension directory was provided and we failed to locate the home "+
			"directory: %s.",
		err,
	)

	// Define the extension's subdirectory name (ex: `~/.vscode/extensions/[here]`)
	extDirName := fmt.Sprintf("%s-%s-%s", publisher, extensionID, version)

	// Look for a `.vscode-oss` or `.vscode` directory in the user's home path
	var extDir string
	extDirOSS := filepath.Join(home, ".vscode-oss")
	extDirMS := filepath.Join(home, ".vscode")
	if _, err := os.Lstat(extDirOSS); err == nil {
		extDir = filepath.Join(extDirOSS, "extensions", extDirName)
		echo.Debugf("Using extension directory [%s].", extDirOSS)

	} else if _, err := os.Lstat(extDirMS); err == nil {
		extDir = filepath.Join(extDirMS, "extensions", extDirName)
		echo.Debugf("Using extension directory [%s].", extDirMS)

	} else {
		echo.Debugf("Attempted: %s.", extDirOSS)
		echo.Debugf("Attempted: %s.", extDirMS)
		echo.Fatal("An extension directory was not provided and we failed to locate one.")
	}

	// Create the extension dir if needed
	err = os.MkdirAll(extDir, fileModeRWX)
	must(
		err == nil || errors.Is(err, os.ErrExist),
		"Failed to create extension output directory [%s]: %s.",
		extDir, err,
	)

	return extDir
}
