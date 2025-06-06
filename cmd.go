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
)

const (
	fileModeRWX = 0o700
	fileModeRW  = 0o600
	fileFlags   = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

func Exec(cfg *Config) {
	// Init the Gallery
	g := NewGallery(cfg.GalleryScheme, cfg.GalleryHost)

	// Exec the command we received
	switch cfg.Command {
	default:
		echo.Fatal(UsageError("Received unknown command [%s].", cfg.Command))

	case cmdInstall:
		echo.Debugf("Received command [%s].", cfg.Command)
		InstallExtension(g, cfg)

	case cmdDownload:
		echo.Debugf("Received command [%s].", cfg.Command)
		DownloadExtension(g, cfg)
	}
}

func InstallExtension(g Gallery, cfg *Config) {
	// Get the `.vsix` file stream
	stream, err := g.GetExtension(context.Background(), cfg.Publisher, cfg.ExtensionID, cfg.Version)
	assert(err == nil, "Failed to fetch Gallery extension: %s.", err)

	// Init the zip reader
	zr, err := zip.NewReader(stream, stream.Size())
	assert(err == nil, "Failed to init zip reader: %s.", err)

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
		output := filepath.Join(cfg.ExtensionDir, name)
		echo.Debugf("Outputting file [%s] to [%s].", name, output)

		// Create any requisite directories
		err = os.MkdirAll(filepath.Dir(output), fileModeRWX)
		assert(
			err == nil || errors.Is(err, os.ErrExist),
			"Failed to create output directory structure [%s]: %s.",
			filepath.Dir(output), err,
		)

		// Get a readable stream to the zipped file
		src, err := zr.Open(zipFile.Name)
		assert(err == nil, "Failed to read zip archive from response body: %s.", err)

		// Get a writable stream to the on-disk file
		dst, err := os.OpenFile(output, fileFlags, fileModeRW)
		assert(err == nil, "Failed to open output file [%s]: %s.", output, err)
		defer dst.Close()

		// Write the file to disk
		_, err = io.Copy(dst, src)
		assert(err == nil, "Failed to write file [%s] to disk: %s.", output, err)
	}

	echo.Infof(
		"[%s-%s] @ [%s] install complete to [%s].",
		cfg.Publisher, cfg.ExtensionID, cfg.Version, cfg.ExtensionDir,
	)
}

func DownloadExtension(g Gallery, cfg *Config) {
	echo.Infof("Fetching extension [%s-%s] @ [%s].", cfg.Publisher, cfg.ExtensionID, cfg.Version)

	// Fetch the extension
	stream, err := g.GetExtension(context.Background(), cfg.Publisher, cfg.ExtensionID, cfg.Version)
	assert(err == nil, "Failed to fetch extension: %s.", err)

	echo.Debugf(
		"Writing file with file mode [%04o] and flags [%010b].", // The highest bit set for file flags is O_TRUNC @ 1 << 9
		fileModeRW, fileFlags,
	)

	// Get a writable file stream to output the extension
	file, err := os.OpenFile(cfg.Output, fileFlags, fileModeRW)
	assert(err == nil, "Failed to get writable stream to output file: %s.", err)
	defer file.Close()

	// Write the vsix package to file
	n, err := io.Copy(file, stream)
	assert(err == nil, "Failed to write extension contents to disk: %s.", err)
	echo.Debugf("Wrote [%d] bytes to file.", n)
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
