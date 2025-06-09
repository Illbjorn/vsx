package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/illbjorn/echo"
)

func ExtensionDir() (string, error) {
	// Get the home directory path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Look for a `.vscode-oss` or `.vscode` directory in the user's home path
	var extDir string
	extDirOSS := filepath.Join(home, ".vscode-oss")
	extDirMS := filepath.Join(home, ".vscode")
	if _, err := os.Lstat(extDirOSS); err == nil {
		extDir = filepath.Join(extDirOSS, "extensions")
		echo.Debugf("Using extension directory [%s].", extDirOSS)

	} else if _, err := os.Lstat(extDirMS); err == nil {
		extDir = filepath.Join(extDirMS, "extensions")
		echo.Debugf("Using extension directory [%s].", extDirMS)

	} else {
		echo.Debugf("Attempted: %s.", extDirOSS)
		echo.Debugf("Attempted: %s.", extDirMS)
		echo.Fatal("An extension directory was not provided and we failed to locate one.")
	}

	return extDir, nil
}

var (
	ErrNoDot = fmt.Errorf("extension input missing dot separator ('publisher.ID')")
	ErrNoPub = fmt.Errorf("extension input missing publisher ('publisher.ID')")
	ErrNoID  = fmt.Errorf("extension input missing ID ('publisher.ID')")
	ErrAtOOB = fmt.Errorf("ill-formed extension version")
)

func ParseExtension(input string) (extPub, extID, extVer string, err error) {
	const notSet = -1
	// Locate the positions of '.' and '@' within string `input`
	dot := notSet
	at := notSet
	for i := range len(input) {
		if input[i] == '.' && dot == notSet {
			dot = i
		}
		if input[i] == '@' && at == notSet {
			at = i
		}
		if dot > notSet && at > notSet {
			break
		}
	}

	// '.' must appear somewhere in `input`
	if dot == notSet {
		err = ErrNoDot
		return
	}

	// '.' must not appear at the very start of `input`
	if dot == 0 {
		err = ErrNoPub
		return
	}

	// '.' must not appear at the very end of `input`
	if dot == len(input)-1 {
		err = ErrNoID
		return
	}

	// Set publisher
	extPub = input[:dot]

	// Set ID and version
	if at == notSet {
		// If we have no '@', the extension ID is the remainder of the string
		extID = input[dot+1:]

	} else {
		// If we have an '@', the extension ID is ['.'+1:'@'] and version is ['@':]
		//
		// The '@' must not appear at the start or end of `input`
		if at == 0 || at == len(input)-1 || at < dot {
			err = ErrAtOOB
			return
		}
		extID = input[dot+1 : at]
		extVer = input[at+1:]
	}

	return
}
