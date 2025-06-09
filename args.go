package main

import (
	"flag"
	"strings"
)

type Flag = string

const (
	// Flags
	flagExtPub        Flag = "extension-publisher"
	flagExtPubShort   Flag = "p"
	flagExtID         Flag = "extension-id"
	flagExtIDShort    Flag = "id"
	flagExtVer        Flag = "extension-version"
	flagExtVerShort   Flag = "v"
	flagExtDir        Flag = "extension-dir"
	flagExtDirShort   Flag = "xd"
	flagGalleryHost   Flag = "gallery-host"
	flagGalleryScheme Flag = "gallery-scheme"
	flagOS            Flag = "os"
	flagArch          Flag = "arch"
	flagArchShort     Flag = "a"
	flagOutput        Flag = "output"
	flagOutputShort   Flag = "o"
	flagDebug         Flag = "debug"
	flagDebugShort    Flag = "d"
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
	flag.StringVar(&args.ExtensionDir, flagExtDir, "", "")
	flag.StringVar(&args.ExtensionDir, flagExtDirShort, "", "")

	// --os
	flag.StringVar(&args.OS, flagOS, "", "")

	// --arch, -a
	flag.StringVar(&args.Arch, flagArch, "", "")
	flag.StringVar(&args.Arch, flagArchShort, "", "")

	// --version, -v
	flag.StringVar(&args.Version, flagExtVer, "latest", "")
	flag.StringVar(&args.Version, flagExtVerShort, "latest", "")

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
