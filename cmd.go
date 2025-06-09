package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/illbjorn/argv"
	"github.com/illbjorn/echo"
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

func Run(g gallery.Gallery, cfg *Config, cmd argv.Command) error {
	switch cmd.Name {
	case "":
		return fmt.Errorf("received no command")

	default:
		return fmt.Errorf("received unknown command [%s]", cmd)

	case cmdQuery:
		return QueryExtensions(g, cmd)

	case cmdInstall:
		return InstallExtensions(g, cfg.ExtensionDir, cmd)

	case cmdDownload:
		// Discern where to put the downloads
		//
		// We'll prefer the user-provided [--output, -o] flag but fall back to the
		// current working directory if the flag was not provided
		flagOutputValues, ok := cmd.Flag(flagOutput, flagOutputShort)
		var output string
		if !ok {
			var err error
			output, err = os.Getwd()
			if err != nil {
				return fmt.Errorf(
					"no download directory was specified and working directory retrieval failed: %w",
					err,
				)
			}
		} else {
			output = flagOutputValues[0]
		}

		return DownloadExtensions(g, output, cmd)
	}
}

func InstallExtensions(g gallery.Gallery, extDir string, cmd argv.Command) error {
	// If we don't have an extension directory, try to locate one in the home
	// directory
	if extDir == "" {
		var err error
		extDir, err = ExtensionDir()
		if err != nil {
			return fmt.Errorf(
				"received no VSCode extension directory and failed to locate one: %w",
				err,
			)
		}
	}

	spawn, wait := goLimit(5)

	// Process all requested extensions
	var errs = make([]error, len(cmd.Args))
	for i, input := range cmd.Args {
		spawn(func() {
			// Parse the extension input
			pub, id, ver, err := ParseExtension(input)
			if err != nil {
				errs[i] = fmt.Errorf(
					"failed to parse extension input[%s]: %s",
					input, err,
				)
				return
			}
			// If we got no `ver` value, use the default ('latest')
			if ver == "" {
				ver = "latest"
			}

			// Assemble the full output directory name
			extDirName := fmt.Sprintf("%s.%s-%s", pub, id, ver)
			extDir = filepath.Join(extDir, extDirName)

			// Get the `.vsix` file stream
			stream, err := g.GetExtension(context.Background(), pub, id, ver)
			if err != nil {
				errs[i] = fmt.Errorf("failed to fetch gallery extension: %w", err)
				return
			}

			// Init the zip reader
			zr, err := zip.NewReader(stream, stream.Size())
			if err != nil {
				errs[i] = fmt.Errorf("failed to init zip reader: %w", err)
				return
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
					errs[i] = fmt.Errorf(
						"failed to create output directory structure: %w",
						err,
					)
					return
				}

				// Get a readable stream to the zipped file
				src, err := zr.Open(zipFile.Name)
				if err != nil {
					errs[i] = fmt.Errorf(
						"failed to read zipped file[%s]: %w",
						zipFile.Name, err,
					)
					return
				}

				// Get a writable stream to the on-disk file
				dst, err := os.OpenFile(output, fileFlagsOverwrite, fileModeRW)
				if err != nil {
					errs[i] = fmt.Errorf(
						"failed to open output file [%s]: %w",
						output, err,
					)
					return
				}
				defer dst.Close()

				// Write the file to disk
				_, err = io.Copy(dst, src)
				if err != nil {
					errs[i] = fmt.Errorf(
						"failed to write zipped file[%s] to disk: %w",
						output, err,
					)
					return
				}
			}

			echo.Infof(
				"[%s-%s] @ [%s] install complete to [%s].",
				pub, id, ver, extDir,
			)
		})
	}

	// Wait for all workers to complete
	wait()

	return nil
}

// TODO: Download progress?
func DownloadExtensions(g gallery.Gallery, outDir string, cmd argv.Command) error {
	spawn, wait := goLimit(5)

	// Create the output directory if necessary
	if err := os.MkdirAll(outDir, fileModeRWX); err != nil {
		return fmt.Errorf("failed to create output directory[%s]: %w", outDir, err)
	}

	errs := make([]error, len(cmd.Args))
	for i, input := range cmd.Args {
		echo.Infof("Processing [%s].", input)
		spawn(func() {
			// Parse the extension input
			pub, id, ver, err := ParseExtension(input)
			if err != nil {
				errs[i] = fmt.Errorf(
					"failed to parse extension input[%s]: %s",
					input, err,
				)
				return
			}
			// If we got no `ver` value, use the default ('latest')
			if ver == "" {
				ver = "latest"
			}

			// Fetch the extension
			echo.Infof("Fetching extension [%s] by [%s] @ [%s].", id, pub, ver)
			stream, err := g.GetExtension(context.Background(), pub, id, ver)
			if err != nil {
				errs[i] = fmt.Errorf(
					"failed to fetch extension: %w",
					err,
				)
				return
			}

			// Construct the output file path
			outFilePath := filepath.Join(outDir, fmt.Sprintf("%s.%s-%s.vsix", pub, id, ver))

			// Get a writable file stream to output the extension
			file, err := os.OpenFile(outFilePath, fileFlagsOverwrite, fileModeRW)
			if err != nil {
				errs[i] = fmt.Errorf(
					"failed to get writable stream to output file: %w",
					err,
				)
				return
			}
			defer file.Close()

			// Write the vsix package to file
			n, err := io.Copy(file, stream)
			if err != nil {
				errs[i] = fmt.Errorf(
					"failed to write extension content to disk: %w",
					err,
				)
				return
			}

			echo.Debugf("Wrote [%d] bytes to file.", n)
		})
	}

	// Wait for all jobs to complete
	wait()

	return errors.Join(errs...)
}

var (
	ErrQueryFailed = fmt.Errorf("failed extension query")
	colHeaders     = [...]string{
		"Name",
		"Publisher",
		"Install With",
		"Installs",
	}
	colSizes = [...]int{
		25,
		20,
		50,
		8,
	}
	// Ensure colHeaders and colSizes remain reasonably in sync
	_ = colSizes[len(colHeaders)-1]
)

func QueryExtensions(g gallery.Gallery, cmd argv.Command) error {
	if len(cmd.Args) == 0 {
		return UsageError("No query terms received.")
	}

	query := strings.Join(cmd.Args, " ")

	printRow(colHeaders[:]...)

	for meta, err := range g.Query(context.Background(), query) {
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQueryFailed, err)
		}

		printRow(
			meta.DisplayName,
			meta.Publisher.DisplayName,
			fmt.Sprintf("%s.%s@%s", meta.Publisher.Name, meta.Name, meta.Versions[0].Version),
			strconv.FormatFloat(meta.Statistics[0].Value, 'f', 0, 64), // Installs
		)
	}

	return nil
}

const (
	colPadding = "  "
)

func printRow(values ...string) {
	for i := range len(colSizes) {
		value := values[i]
		valueLen := len(value)
		colSize := colSizes[i]
		// TODO: Actually fix the issue with spacing around multi-byte characters
		// and remove this hack
		if len(value) != len([]rune(value)) {
			colSize -= 1
		}

		if valueLen > colSize {
			// If the value is oversized, truncate and write ellipses to indicate
			// truncation
			fmt.Printf("%-*s", colSize, value[:colSize-3]+"...")

		} else if valueLen == colSize {
			// If the value matches exactly, just write it out normally
			fmt.Printf("%-*s", colSize, value)

		} else if valueLen < colSize {
			// If the value is undersized, pad to the right with spaces
			fmt.Printf("%-*s", colSize, value)
		}

		// For all but the final column, pad between columns
		if i < len(colSizes)-1 {
			fmt.Print(colPadding)
		}
	}

	// Finish the row with a newline
	fmt.Println()
}

func UsageError(msg string, values ...any) error {
	msg = fmt.Sprintf(msg, values...)
	return fmt.Errorf("%s\n%s", msg, Usage())
}

func Usage() string {
	return `
>> Overview

  VSX is a simple (in-progress) command-line VSCode extension manager.

  To get off the ground quickly, refer to the quickstart:
  https://github.com/illbjorn/vsx

>> Usage

  vsx [install [EXTENSION] | download [EXTENSION] | query [TERMS]] [FLAGS] 
                ┗━━━┳━━━┛              ┗━━━┳━━━┛
                    ┃                      ┃
  ┏━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━┛
  ┃ Example
  ┣━
  ┃ usernamehw.errorlens@3.26.0
  ┣━
  ┃ usernamehw -> Extension Publisher
  ┃  errorlens -> Extension ID
  ┃    @3.26.0 -> Optional, allows specific version extension installation
  ┃               If not provided, a default of 'latest' will be used
  ┗━

>> Commands

   install   Download an extension and install it.
   download  Download the extension and output the .vsix file to disk.
   query     Query the extension catalog.

>> Flags

  --extension-dir, -xd  The local file path to your
                        '.vscode/extensions' directory.
                        Default:
                        1. ~/.vscode-oss/extensions
                        2. ~/.vscode/extensions
  --gallery-scheme      The URI scheme for requests to the Gallery
                        ('HTTP' or 'HTTPS').
                        Default: HTTPS
  --gallery-host        The hostname of the extension Gallery
                        (example: my.gallery.com).
  --output,        -o   If the command provided is 'download', '--output' is 
                        where the .vsix package will be saved. 
                        Default: './[publisherID]-[extensionID].[version].vsix'
  --debug,         -d   Enables additional logging for troubleshooting
                        purposes.

>> Environment Variables

  To avoid giant run-on commands, VSX supports environment variables for the
  primary values required by every command.

  ┏━
  ┃ NOTE: Provided flag values will supersede values identified in the
  ┃ environment!
  ┗━

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
