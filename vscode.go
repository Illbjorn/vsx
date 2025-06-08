package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illbjorn/echo"
)

func ExtensionDir(extPub, extID, extVer string) (string, error) {
	// Get the home directory path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Define the extension's subdirectory name (ex: `~/.vscode/extensions/[here]`)
	extDirName := fmt.Sprintf("%s-%s-%s", extPub, extID, extVer)

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
	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf(
			"failed to create extension output directory[%s]: %w",
			extDir, err,
		)
	}

	return extDir, nil
}
