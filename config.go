package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func LoadConfig() *Config {
	cfg := new(Config)

	// First order of precedence is command-line args
	loadConfigArgs(cfg)
	// Anything left empty we fill in with environment variable values where
	// possible
	loadConfigEnv(cfg)
	// Finally, apply defaults wherever possible for any remaining empty values
	applyConfigDefaults(cfg)
	// Assert requirements
	validateConfig(cfg)

	return cfg
}

type Config struct {
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

	// Command is the selected subcommand (ex: `install`, `download`)
	Command string

	// Publisher is the VSIX extension publisher
	Publisher string

	// ExtensionID is the unique extension identifier string
	ExtensionID string

	Debug bool
}

func loadConfigEnv(cfg *Config) {
	const envGalleryHost = "VSX_GALLERY_HOST"
	if cfg.GalleryHost == "" {
		cfg.GalleryHost = os.Getenv(envGalleryHost)
	}

	const envGalleryScheme = "VSX_GALLERY_SCHEME"
	if cfg.GalleryScheme == "" {
		cfg.GalleryScheme = os.Getenv(envGalleryScheme)
	}

	const envExtensionDir = "VSX_EXTENSION_DIR"
	if cfg.ExtensionDir == "" {
		cfg.ExtensionDir = os.Getenv(envExtensionDir)
	}

	const envOS = "VSX_OS"
	if cfg.OS == "" {
		cfg.OS = os.Getenv(envOS)
	}

	const envArch = "VSX_ARCH"
	if cfg.Arch == "" {
		cfg.Arch = os.Getenv(envArch)
	}
}

func loadConfigArgs(cfg *Config) {
	// Define flags
	//
	// --gallery-host
	flag.StringVar(&cfg.GalleryHost, "gallery-host", "", "")

	// --gallery-scheme
	flag.StringVar(&cfg.GalleryScheme, "gallery-scheme", "", "")

	// --extension-dir, -xd
	flag.StringVar(&cfg.ExtensionDir, "extension-dir", "", "")
	flag.StringVar(&cfg.ExtensionDir, "xd", "", "")

	// --os
	flag.StringVar(&cfg.OS, "os", "", "")

	// --arch, -a
	flag.StringVar(&cfg.Arch, "arch", "", "")
	flag.StringVar(&cfg.Arch, "a", "", "")

	// --version, -v
	flag.StringVar(&cfg.Version, "version", "latest", "")
	flag.StringVar(&cfg.Version, "v", "latest", "")

	// --output, -o
	flag.StringVar(&cfg.Output, "output", "", "")
	flag.StringVar(&cfg.Output, "o", "", "")

	// --debug, -d
	flag.BoolVar(&cfg.Debug, "debug", false, "")
	flag.BoolVar(&cfg.Debug, "d", false, "")

	// Parse
	flag.Parse()
	positionalArgs := flag.Args()

	// Assign the command
	if len(positionalArgs) > 0 {
		cfg.Command = strings.ToLower(positionalArgs[0])
	}

	// Capture the publisher and extension IDs
	//
	// These are provided like `[publisherID]-[extensionID]` as the last
	// positional arg
	identifier := positionalArgs[len(positionalArgs)-1]
	cfg.Publisher, cfg.ExtensionID, _ = strings.Cut(identifier, ".")
}

func applyConfigDefaults(cfg *Config) {
	// If we didn't find a gallery scheme, use `https`
	if cfg.GalleryScheme == "" {
		cfg.GalleryScheme = "https"
	}

	// If we didn't receive a path to save the vsix package to, default to
	// `[publisher]-[extensionID]-[version].vsix` in the current working directory
	if cfg.Command == cmdDownload && cfg.Output == "" {
		cfg.Output = fmt.Sprintf("%s-%s-%s.vsix", cfg.Publisher, cfg.ExtensionID, cfg.Version)
	}

	// If we don't have an extension directory, try to locate one in the home
	// directory
	if cfg.Command == cmdInstall && cfg.ExtensionDir == "" {
		cfg.ExtensionDir = ExtensionDir(cfg.Publisher, cfg.ExtensionID, cfg.Version)
	}
}

func validateConfig(cfg *Config) {
	// Command must be set
	assert(cfg.Command != "", UsageError(
		"Received no command (ex: `download`, `install`).",
	))

	// ExtensionID must be set
	assert(cfg.ExtensionID != "", UsageError(
		"Received no extension ID.",
	))

	// Publisher must be set
	assert(cfg.Publisher != "", UsageError(
		"Received no extension publisher.",
	))
}
