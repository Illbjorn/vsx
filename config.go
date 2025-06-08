package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	cfgFileName = "vsx.json"
)

var (
	ErrConfigDir    = fmt.Errorf("failed to identify the user config dir")
	ErrConfigPath   = fmt.Errorf("failed to retrieve the config file path")
	ErrOpenConfig   = fmt.Errorf("failed to open config file")
	ErrDecodeConfig = fmt.Errorf("failed to decode config file")
	ErrEncodeConfig = fmt.Errorf("failed to encode config file")
	ErrMkCfgDir     = fmt.Errorf("failed to create the configuration dir")
)

func LoadConfigFile() (*Config, error) {
	cfg := new(Config)

	// Attempt to load the config from file
	//
	// Get the full config file path
	path, err := cfgFile(cfgFileName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigPath, err)
	}

	// Get a readable stream to the config file
	f, err := os.OpenFile(path, fileFlagsRead, fileModeRW)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenConfig, err)
	}
	defer f.Close()

	// Attempt the decode
	err = json.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeConfig, err)
	}

	// For any values left blank, apply reasonable defaults
	cfg = applyConfigDefaults(cfg)

	return cfg, nil
}

func SaveConfigFile(cfg *Config) error {
	// Create the config directory if necessary
	cfgDir, err := cfgRoot()
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			ErrConfigDir, err,
		)
	}
	if err := os.MkdirAll(cfgDir, fileModeRWX); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %w", ErrMkCfgDir, err)
	}

	// Get the full config file path
	cfgPath, err := cfgFile(cfgFileName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfigPath, err)
	}

	// Get a writable stream to the config file
	f, err := os.OpenFile(cfgPath, fileFlagsOverwrite, fileModeRW)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenConfig, err)
	}
	defer f.Close()

	// Attempt the encode
	err = json.NewEncoder(f).Encode(cfg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEncodeConfig, err)
	}

	return nil
}

func LoadConfigEnv(cfg *Config) *Config {
	const envGalleryHost = "VSX_GALLERY_HOST"
	if v, ok := os.LookupEnv(envGalleryHost); ok {
		cfg.GalleryHost = v
	}

	const envGalleryScheme = "VSX_GALLERY_SCHEME"
	if v, ok := os.LookupEnv(envGalleryScheme); ok {
		cfg.GalleryScheme = v
	}

	const envExtensionDir = "VSX_EXTENSION_DIR"
	if v, ok := os.LookupEnv(envExtensionDir); ok {
		cfg.ExtensionDir = v
	}

	const envOS = "VSX_OS"
	if v, ok := os.LookupEnv(envOS); ok {
		cfg.OS = v
	}

	const envArch = "VSX_ARCH"
	if v, ok := os.LookupEnv(envArch); ok {
		cfg.Arch = v
	}

	return cfg
}

type Config struct {
	// ExtensionDir is utilized by the `install` subcommand and refers to the
	// `.vscode` or `.vscode-oss` extensions directory
	ExtensionDir string `json:"extensions_dir"`

	// GalleryScheme pairs with GalleryHost and refers to the URI scheme for
	// use in communication with the supplied extension gallery
	GalleryScheme string `json:"gallery_scheme"`

	// GalleryHost pairs with GalleryScheme and refers to the hostname of a given
	// extension gallery
	GalleryHost string `json:"gallery_host"`

	// OS is the targeted extension operating system
	OS string `json:"os"`

	// Arch is the targeted extension architecture
	Arch string `json:"arch"`

	// HistFilePath is the path to the history file for REPL command history
	HistFilePath string `json:"hist_file_path"`
}

func applyConfigDefaults(cfg *Config) *Config {
	// If we don't have a gallery scheme, use `https`
	if cfg.GalleryScheme == "" {
		cfg.GalleryScheme = "https"
	}
	// If we don't have a history file path, use `.history` alongside the config
	// file
	if cfg.HistFilePath == "" {
		cfg.HistFilePath, _ = cfgFile(".history")
	}

	return cfg
}

// cfgFile produces a path to file `name` within the VSX-specific root
// configuration directory
func cfgFile(name string) (string, error) {
	root, err := cfgRoot()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrConfigDir, err)
	}
	return filepath.Join(root, name), nil
}

// cfgRoot produces the VSX-specific root configuration directory
func cfgRoot() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return cfgDir, fmt.Errorf("%w: %w", ErrConfigDir, err)
	}
	return filepath.Join(cfgDir, app), nil
}

// MergeInputs merges any values which can be considered persistent
// configuration from the command-line inputs into the configuration itself
func MergeInputs(cfg *Config, flags map[string][]string) *Config {
	if v, ok := flags[flagExtDir]; ok {
		cfg.ExtensionDir = v[0]
	}

	if v, ok := flags[flagGalleryScheme]; ok {
		cfg.GalleryScheme = v[0]
	}

	if v, ok := flags[flagGalleryHost]; ok {
		cfg.GalleryHost = v[0]
	}

	if v, ok := flags[flagOS]; ok {
		cfg.OS = v[0]
	}

	if v, ok := flags[flagArch]; ok {
		cfg.Arch = v[0]
	}

	return cfg
}
