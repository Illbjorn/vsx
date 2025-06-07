package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/illbjorn/echo"
)

type CMD = string

const (
	// CMDs
	cmdQuery    CMD = "query"
	cmdInstall  CMD = "install"
	cmdDownload CMD = "download"
	cmdExit     CMD = "exit"
)

const (
	fileModeRWX = 0o700
	fileModeRW  = 0o600
	fileFlags   = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

func RunCMD(g Gallery, cfg *Config, cmd string, flags map[string][]string, args []string) error {
	// Exec the command we received
	switch cmd {
	default:
		return fmt.Errorf("received unknown command [%s]", cmd)

	case cmdInstall:
		echo.Debugf("Received command [%s].", cmd)

		// Collect required inputs

		publishers, ok := flags[flagPublisher]
		if !ok {
			return errMissingInput(cmd, flagPublisher)
		}

		extensionIDs, ok := flags[flagExtensionID]
		if !ok {
			errMissingInput(cmd, flagExtensionID)
		}

		versions, ok := flags[flagVersion]
		if !ok {
			errMissingInput(cmd, flagVersion)
		}

		// Install one or more extensions
		//
		// TODO: Currently this takes the shortest length of the three required
		// values and only performs the install for that series of information. This
		// should be updated to produce a clear error around which set was missing
		// which bit of required information.
		var errs error
		for i := range min(len(publishers), len(extensionIDs), len(versions)) {
			err := InstallExtension(
				g,
				publishers[i],
				extensionIDs[i],
				versions[i],
				cfg.ExtensionDir,
			)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf(
					"failed to install extension [%s-%s]: %w",
					publishers[i], extensionIDs[i], err,
				))
			}
		}
		return errs

	case cmdDownload:
		echo.Debugf("Received command [%s].", cmd)

		// Collect required inputs

		publishers, ok := flags[flagPublisher]
		if !ok {
			return errMissingInput(cmd, flagPublisher)
		}

		extensionIDs, ok := flags[flagExtensionID]
		if !ok {
			return errMissingInput(cmd, flagExtensionID)
		}

		versions, ok := flags[flagVersion]
		if !ok {
			return errMissingInput(cmd, flagVersion)
		}

		// Install one or more extensions
		//
		// TODO: Currently this takes the shortest length of the three required
		// values and only performs the install for that series of information. This
		// should be updated to produce a clear error around which set was missing
		// which bit of required information.
		var errs error
		for i := range min(len(publishers), len(extensionIDs), len(versions)) {
			err := DownloadExtension(
				g,
				publishers[i],
				extensionIDs[i],
				versions[i],
				cfg.ExtensionDir,
			)
			if err != nil {
				errors.Join(errs, fmt.Errorf(
					"failed to download extension [%s-%s]: %w",
					publishers[i], extensionIDs[i], err,
				))
			}
		}
		return errs
	}
}

func errMissingInput(cmd string, input string) error {
	return fmt.Errorf(
		"Missing required input [%s] for command [%s].",
		input, cmd,
	)
}

var prefix = readline.PcItem
var completer = readline.NewPrefixCompleter(
	// query
	prefix("query", prefix("--name")),
	// download
	prefix("download", prefix("--name")),
	// install
	prefix("install", prefix("--name")),
	// backup
	prefix("backup", prefix("--to")),
)

func EnterREPL(g Gallery, cfg *Config) {
	// Init the readline instance
	l, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[31m»\033[0m ",
		HistoryFile:  cfg.HistFilePath,
		AutoComplete: completer,
	})
	must(err == nil, "Failed to init readline instance: %s.", err)

	for {
		// Wait for a command...
		v, err := l.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return
			}
			echo.Fatalf("Readline failed: %s.", err)
		}

		// Tokenize the command
		tokens := tokenizeCMD(v)
		if len(tokens) == 0 {
			continue
		}

		// Parse the tokenize command
		cmd, flags, args := parseCMD(tokens)
		_ = args

		// Exec the command as necessary
		switch cmd {
		default:
			echo.Errorf("Received missing or unknown command [%s].", cmd)
			continue

		case cmdDownload:
			for _, toDownload := range flags["name"] {
				echo.Infof("Downloading extension [%s].", toDownload)
			}

		case cmdInstall:
			for _, toInstall := range flags["name"] {
				echo.Infof("Installing extension [%s].", toInstall)
				// TODO: Figure the fuck out how I'm going to handle piping these fields
				// (publisher, extensionID, version, extensionDir) all over god's green
				// fucking earth because this API is fucking dogshit.
			}

		case cmdQuery:

		case cmdExit:
			return
		}
	}
}

func InstallExtension(g Gallery, publisher, extensionID, version, extensionDir string) error {
	// Get the `.vsix` file stream
	stream, err := g.GetExtension(context.Background(), publisher, extensionID, version)
	if err != nil {
		return fmt.Errorf("failed to fetch gallery extension: %w", err)
	}

	// Init the zip reader
	zr, err := zip.NewReader(stream, stream.Size())
	if err != nil {
		return fmt.Errorf("failed to init zip reader: %w", err)
	}

	// Unzip all files
	for _, zipFile := range zr.File {
		// Ignore non-`extension`-directory files
		if !strings.HasPrefix(zipFile.Name, "extension") {
			echo.Debugf("Skipping file [%s].", zipFile.Name)
			continue
		}

		// Slice off the `extension` prefix
		i := strings.IndexByte(zipFile.Name, '/')
		if i == -1 || i == len(zipFile.Name)-1 {
			echo.Debugf("Skipping zipped file [%s](no path suffix).", zipFile.Name)
			continue
		}
		name := zipFile.Name[i+1:]

		// Define the output path
		output := filepath.Join(extensionDir, name)
		echo.Debugf("Outputting file [%s] to [%s].", name, output)

		// Create any requisite directories
		err = os.MkdirAll(filepath.Dir(output), fileModeRWX)
		if err != nil {
			return fmt.Errorf("failed to create output directory structure: %w", err)
		}

		// Get a readable stream to the zipped file
		src, err := zr.Open(zipFile.Name)
		if err != nil {
			return fmt.Errorf("failed to read zip archive from response body: %w", err)
		}

		// Get a writable stream to the on-disk file
		dst, err := os.OpenFile(output, fileFlags, fileModeRW)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer dst.Close()

		// Write the file to disk
		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("failed to write file [%s] to disk: %w", output, err)
		}
	}

	echo.Infof(
		"[%s-%s] @ [%s] install complete to [%s].",
		publisher, extensionID, version, extensionDir,
	)

	return nil
}

func DownloadExtension(g Gallery, publisher, extensionID, version, output string) error {
	echo.Infof("Fetching extension [%s-%s] @ [%s].", publisher, extensionID, version)

	// Fetch the extension
	stream, err := g.GetExtension(context.Background(), publisher, extensionID, version)
	if err != nil {
		return fmt.Errorf("failed to fetch extension: %w", err)
	}

	echo.Debugf(
		"Writing file with file mode [%04o] and flags [%010b].", // The highest bit set for file flags is O_TRUNC @ 1 << 9
		fileModeRW, fileFlags,
	)

	// Get a writable file stream to output the extension
	file, err := os.OpenFile(output, fileFlags, fileModeRW)
	if err != nil {
		return fmt.Errorf("failed to get writable stream to output file: %w", err)
	}
	defer file.Close()

	// Write the vsix package to file
	n, err := io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("failed to write extension content to disk: %w", err)
	}

	echo.Debugf("Wrote [%d] bytes to file.", n)

	return nil
}

func UsageError(msg string, values ...any) string {
	return fmt.Sprintf(msg, values...) + "\n" + Usage()
}

func Usage() string {
	return `
>> Overview

   VSX is a simple (in-progress) command-line VSCode extension manager.

	 To get off the ground quickly, refer to the quickstart:
	 https://github.com/illbjorn/vsx

>> Usage

   vsx [FLAGS] [COMMAND] [EXTENSION]

   * NOTE: '[EXTENSION]' is the '[publisherID-extensionID]' component of a
   * Gallery item. These values can be found in the pre-populated 'ext install'
   * command on a Gallery extension's page or from the 'itemName' query
   * parameter when browsing the marketplace.
   * Example: items?itemName=modular-mojotools.vscode-mojo
   *                         ^---------------------------^

>> Commands

   install   Download an extension and install it.
   download  Download the extension and output the .vsix file to disk.

>> Flags

   --extension-dir, -xd  The local file path to your '.vscode/extensions'
                         directory.
                         Default:
                           1. ~/.vscode-oss/extensions
                           2. ~/.vscode/extensions
   --gallery-scheme      The URI scheme for requests to the Gallery ('HTTP' or
                         'HTTPS').
                         Default: HTTPS
   --gallery-host        The hostname of the extension Gallery (example:
                         my.gallery.com).
   --version,       -v   The version of the extension to install
                         Default: 'latest'
   --output,        -o   If the command provided is 'download', '--output' is
                         where the .vsix package will be saved.
                         Default: './[publisherID]-[extensionID].[version].vsix'
   --debug,         -d   Enables additional logging for troubleshooting
                         purposes.

>> Environment Variables

   To avoid giant run-on commands, VSX supports environment variables for the
   primary values required by every command.

   * NOTE: Provided flag values will supersede values identified in the
   * environment!

   VSX_GALLERY_HOST    The hostname of the extension Gallery (example:
                       my.gallery.com).
                       Flag: --gallery-host

   VSX_GALLERY_SCHEME  The URI scheme for requests to the Gallery ('HTTP' or
                       'HTTPS').
                       Flag: --gallery-scheme

   VSX_EXTENSION_DIR   The local file path to your '.vscode/extensions'
                       directory.
                       Flag: --extension-dir, -xd
`
}
