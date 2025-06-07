package main

import (
	"flag"
	"fmt"
	"strings"
)

type Flag = string

const (
	// Flags
	flagPublisher         Flag = "publisher"
	flagExtensionID       Flag = "extension-id"
	flagGalleryHost       Flag = "gallery-host"
	flagGalleryScheme     Flag = "gallery-scheme"
	flagExtensionDir      Flag = "extension-dir"
	flagExtensionDirShort Flag = "xd"
	flagOS                Flag = "os"
	flagArch              Flag = "arch"
	flagArchShort         Flag = "a"
	flagVersion           Flag = "version"
	flagVersionShort      Flag = "v"
	flagOutput            Flag = "output"
	flagOutputShort       Flag = "o"
	flagDebug             Flag = "debug"
	flagDebugShort        Flag = "d"
)

func ParseArgs() *Args {
	args := new(Args)

	// Define flags
	//
	// --gallery-host
	flag.StringVar(&args.GalleryHost, flagGalleryHost, "", "")

	// --gallery-scheme
	flag.StringVar(&args.GalleryScheme, flagGalleryScheme, "", "")

	// --extension-dir, -xd
	flag.StringVar(&args.ExtensionDir, flagExtensionDir, "", "")
	flag.StringVar(&args.ExtensionDir, flagExtensionDirShort, "", "")

	// --os
	flag.StringVar(&args.OS, flagOS, "", "")

	// --arch, -a
	flag.StringVar(&args.Arch, flagArch, "", "")
	flag.StringVar(&args.Arch, flagArchShort, "", "")

	// --version, -v
	flag.StringVar(&args.Version, flagVersion, "latest", "")
	flag.StringVar(&args.Version, flagVersionShort, "latest", "")

	// --output, -o
	flag.StringVar(&args.Output, flagOutput, "", "")
	flag.StringVar(&args.Output, flagOutputShort, "", "")

	// --debug, -d
	flag.BoolVar(&args.Debug, flagDebug, false, "")
	flag.BoolVar(&args.Debug, flagDebugShort, false, "")

	// Parse
	flag.Parse()
	positionalArgs := flag.Args()

	// Assign the command
	if len(positionalArgs) > 0 {
		args.Command = strings.ToLower(positionalArgs[0])
		// Capture the publisher and extension IDs
		//
		// These are provided like `[publisherID]-[extensionID]` as the last
		// positional arg
		identifier := positionalArgs[len(positionalArgs)-1]
		args.Publisher, args.ExtensionID, _ = strings.Cut(identifier, ".")
	}

	return args
}

type Args struct {
	// Command is the selected subcommand (ex: `install`, `download`)
	Command string

	// Publisher is the VSIX extension publisher
	Publisher string

	// ExtensionID is the unique extension identifier string
	ExtensionID string

	// ExtensionDir is utilized by the `install` subcommand and refers to the
	// `.vscode` or `.vscode-oss` extensions directory
	ExtensionDir string

	// Output is utilized by the `download` subcommand and refers to the full
	// file path to which the extension `vsix` file will be downloaded.
	Output string

	// GalleryScheme pairs with GalleryHost and refers to the URI scheme for
	// use in communication with the supplied extension gallery
	GalleryScheme string

	// GalleryHost pairs with GalleryScheme and refers to the hostname of a given
	// extension gallery
	GalleryHost string

	// OS is the targeted extension operating system
	OS string

	// Arch is the targeted extension architecture
	Arch string

	// Version is the desired extension version
	Version string

	Debug bool
}

func applyArgDefaults(args *Args) *Args {
	// If we didn't receive a path to save the vsix package to, default to
	// `[publisher]-[extensionID]-[version].vsix` in the current working directory
	if args.Command == cmdDownload && args.Output == "" {
		args.Output = fmt.Sprintf("%s-%s-%s.vsix", args.Publisher, args.ExtensionID, args.Version)
	}

	// If we don't have an extension directory, try to locate one in the home
	// directory
	if args.Command == cmdInstall && args.ExtensionDir == "" {
		args.ExtensionDir = ExtensionDir(args.Publisher, args.ExtensionID, args.Version)
	}

	return args
}
