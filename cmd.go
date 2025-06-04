package main

import (
	"archive/zip"
	"context"
	"errors"
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

func InstallExtension(g Gallery, installTo, publisherID, extensionID, version string) {
	// Get the `.vsix` file stream
	stream, err := g.GetExtension(context.Background(), publisherID, extensionID, version)
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
		output := filepath.Join(installTo, name)
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
		publisherID, extensionID, version, installTo,
	)
}

func DownloadExtension(g Gallery, output, publisher, extensionID, version string) {
	echo.Infof("Fetching extension [%s-%s] @ [%s].", publisher, extensionID, version)

	// Fetch the extension
	stream, err := g.GetExtension(context.Background(), publisher, extensionID, version)
	assert(err == nil, "Failed to fetch extension: %s.", err)

	echo.Debugf(
		"Writing file with file mode [%04o] and flags [%010b].", // The highest bit set for file flags is O_TRUNC @ 1 << 9
		fileModeRW, fileFlags,
	)

	// Get a writable file stream to output the extension
	file, err := os.OpenFile(output, fileFlags, fileModeRW)
	assert(err == nil, "Failed to get writable stream to output file: %s.", err)
	defer file.Close()

	// Write the vsix package to file
	n, err := io.Copy(file, stream)
	assert(err == nil, "Failed to write extension contents to disk: %s.", err)
	echo.Debugf("Wrote [%d] bytes to file.", n)
}
