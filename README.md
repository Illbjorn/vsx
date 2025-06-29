
# Overview

`vsix` is a basic CLI tool to facilitate VS Code `.vsix` package downloads from
a [VSCode Extension Gallery service](https://github.com/microsoft/vscode/blob/main/src/vs/platform/extensionManagement/common/extensionGalleryService.ts). 

# Quickstart

## Install `VSX`

### Via `go install`

```bash
go install github.com/illbjorn/vsx@latest
```

### From Releases

#### TODO

## Set the Gallery Hostname

```bash
export VSX_HOST="example.gallery.com"
```

## Install a package

Locate your package identifier. This is the bit following `ext install` in 
the auto-generated install snippet in the standard Microsoft Gallery 
([example](https://marketplace.visualstudio.com/items?itemName=modular-mojotools.vscode-mojo)).

```bash
vsx install modular-mojotools.vscode-mojo
```

## Download a Package

```bash
vsx download -o mojo.vsix modular-mojotools.vscode-mojo
```

# Usage

```bash
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
```

# TODO

```bash
- TODO: Implement signature verification of downloaded VSIX files (PKCS #1 / v1.5)
- TODO: Implement `config` subcommand to manually persist configuration values
- TODO: Implement `update` subcommand
- TODO: Implement `backup` and `restore` subcommands
- TODO: Implement `list` subcommand
- TODO: Implement timeout support (init contexts, pass with timeout to CMD handlers)
```
