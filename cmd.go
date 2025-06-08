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

	"github.com/illbjorn/echo"
	"github.com/illbjorn/vsx/argv"
	"github.com/illbjorn/vsx/gallery"
)

type CMD = string

const (
	// CMDs
	cmdQuery    CMD = "query"
	cmdInstall  CMD = "install"
	cmdDownload CMD = "download"
	cmdList     CMD = "list"
	cmdExit     CMD = "exit"
)

func RunCMD(g gallery.Gallery, cfg *Config, cmd argv.Command) error {
	switch cmd.Name {
	case "":
		return fmt.Errorf("received no command")

	default:
		return fmt.Errorf("received unknown command [%s]", cmd)

	case cmdInstall:
		echo.Debugf("Received command [%s].", cmd)

		// Collect required inputs

		publishers, ok := cmd.Flags[flagExtPub]
		if !ok {
			return errMissingInput(cmd.Name, flagExtPub)
		}

		extensionIDs, ok := cmd.Flags[flagExtID]
		if !ok {
			return errMissingInput(cmd.Name, flagExtID)
		}

		versions, ok := cmd.Flags[flagExtVer]
		if !ok {
			return errMissingInput(cmd.Name, flagExtVer)
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

		publishers, ok := cmd.Flags[flagExtPub]
		if !ok {
			return errMissingInput(cmd.Name, flagExtPub)
		}

		extensionIDs, ok := cmd.Flags[flagExtID]
		if !ok {
			return errMissingInput(cmd.Name, flagExtID)
		}

		versions, ok := cmd.Flags[flagExtVer]
		if !ok {
			return errMissingInput(cmd.Name, flagExtVer)
		}

		// Install one or more extensions
		//
		// TODO: Currently this takes the shortest length of the three required
		// values and only performs the install for that series of information.
		// This should be updated to produce a clear error around which set was
		// missing which bit of required information.
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

	case cmdQuery:
		err := QueryExtensions(g, cmd.Args...)
		if err != nil {

		}
		return nil
	}
}

func errMissingInput(cmd string, input string) error {
	return fmt.Errorf(
		"Missing required input [%s] for command [%s].",
		input, cmd,
	)
}

func InstallExtension(g gallery.Gallery, extDir, extPublisher, extID, extVersion string) error {
	// If we don't have an extension directory, try to locate one in the home
	// directory
	if extDir == "" {
		var err error
		extDir, err = ExtensionDir(
			extPublisher,
			extID,
			extVersion,
		)
		if err != nil {
			echo.Fatalf(
				"failed to identify a VSCode extension directory: %w",
				err,
			)
		}
	}

	// Get the `.vsix` file stream
	stream, err := g.GetExtension(context.Background(), extPublisher, extID, extVersion)
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
		output := filepath.Join(extDir, name)
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
		dst, err := os.OpenFile(output, fileFlagsOverwrite, fileModeRW)
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
		extPublisher, extID, extVersion, extDir,
	)

	return nil
}

func DownloadExtension(g gallery.Gallery, extPublisher, extID, extVersion, outputPath string) error {
	echo.Infof("Fetching extension [%s-%s] @ [%s].", extPublisher, extID, extVersion)

	// Fetch the extension
	stream, err := g.GetExtension(context.Background(), extPublisher, extID, extVersion)
	if err != nil {
		return fmt.Errorf("failed to fetch extension: %w", err)
	}

	echo.Debugf(
		"Writing file with file mode [%04o] and flags [%010b].", // The highest bit set for file flags is O_TRUNC @ 1 << 9
		fileModeRW, fileFlagsOverwrite,
	)

	// Get a writable file stream to output the extension
	file, err := os.OpenFile(outputPath, fileFlagsOverwrite, fileModeRW)
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

var (
	ErrQueryFailed = fmt.Errorf("failed extension query")
)

func QueryExtensions(g gallery.Gallery, terms ...string) error {
	if len(terms) == 0 {
		UsageError("No query terms received.")
		return nil
	}

	query := strings.Join(terms, " ")

	echo.Infof(
		"  %25s  %-10s  %-10s  %-20s  %-20s",
		"Name", "Version", "Validated?", "Publisher", "Last Updated",
	)
	echo.Infof(
		"  %25s  %-10s  %-10s  %-20s  %-20s",
		"----", "-------", "----------", "---------", "------------",
	)
	for meta, err := range g.Query(context.Background(), query) {
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQueryFailed, err)
		}
		// displayName  version  validated?  author  lastUpdated
		echo.Infof(
			"  %25s  %-10s  %-10t  %-20s  %-20s",
			meta.DisplayName,
			meta.Versions[0].Version,
			meta.Versions[0].Flags == "validated",
			meta.Publisher.DisplayName,
			meta.LastUpdated.Format("2006-01-02 03:04"),
		)
	}

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
	 query     Query the extension catalog.

>> Flags

   --extension-dir,          -xd  The local file path to your
	                                '.vscode/extensions' directory.
                                  Default:
                                    1. ~/.vscode-oss/extensions
                                    2. ~/.vscode/extensions
   --gallery-scheme               The URI scheme for requests to the Gallery
	                                ('HTTP' or 'HTTPS').
                                  Default: HTTPS
   --gallery-host                 The hostname of the extension Gallery
	                                (example: my.gallery.com).
   --extension-publisher-id  -id  The extension publisher, as reflected in the
	                                Gallery.
   --extension-version,      -v   The version of the extension to install
                                  Default: 'latest'
	 --extension-id,           -id  The extension name, as reflected in the
	                                Gallery.
   --output,                 -o   If the command provided is 'download',
	                                '--output' is where the .vsix package will be
																	saved.
                                  Default: './[publisherID]-[extensionID].[version].vsix'
   --debug,                  -d   Enables additional logging for troubleshooting
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
