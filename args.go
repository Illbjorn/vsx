package main

import (
	"flag"
	"os"
)

type Args struct {
	ExtensionDir  string
	GalleryScheme string
	GalleryHost   string
	OS            string
	Arch          string
	Version       string
	Output        string
	Debug         bool
	Positional    []string
}

func ParseArgs() Args {
	var args Args

	// Define flags
	//
	// --gallery-host
	const envGalleryHost = "VSX_GALLERY_HOST"
	flag.StringVar(&args.GalleryHost, "gallery-host", os.Getenv(envGalleryHost), "")

	// --gallery-scheme
	const envGalleryScheme = "VSX_GALLERY_SCHEME"
	flag.StringVar(&args.GalleryScheme, "gallery-scheme", os.Getenv(envGalleryScheme), "")

	// --extension-dir, -xd
	const envExtensionDir = "VSX_EXTENSION_DIR"
	flag.StringVar(&args.ExtensionDir, "extension-dir", os.Getenv(envExtensionDir), "")
	flag.StringVar(&args.ExtensionDir, "xd", os.Getenv(envExtensionDir), "")

	// --os
	const envOS = "VSX_OS"
	flag.StringVar(&args.OS, "os", os.Getenv(envOS), "")

	// --arch, -a
	const envArch = "VSX_ARCH"
	flag.StringVar(&args.Arch, "arch", os.Getenv(envArch), "")
	flag.StringVar(&args.Arch, "a", os.Getenv(envArch), "")

	// --version, -v
	flag.StringVar(&args.Version, "version", "", "")
	flag.StringVar(&args.Version, "v", "", "")

	// --output, -o
	flag.StringVar(&args.Output, "output", "", "")
	flag.StringVar(&args.Output, "o", "", "")

	// --debug, -d
	flag.BoolVar(&args.Debug, "debug", false, "")
	flag.BoolVar(&args.Debug, "d", false, "")

	// Parse
	flag.Parse()
	args.Positional = flag.Args()
	assert(len(args.Positional) >= 2, Usage())

	return args
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
