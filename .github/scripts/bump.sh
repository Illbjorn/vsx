#!/usr/bin/env bash

: '

>> Overview

  Performs a bump of the provided version segment ('major', 'minor' or 'patch')
  to the version string located in file `.github/VERSION`.

  The above version string represents the version of the Go module and is used
  when performing Git tag creation and pushing 'release' refs (`vX.X.X`).

>> Inputs

  All inputs are injected into the environment by Task (see `taskfile.yml`) at
  the repository root.

  ver_path    : REQUIRED
    The path to the `VERSION` file, in which the version string should be
    updated.

  segment : REQUIRED
    The version segment to be incremented. May be one of: 'major', 'minor' or
    'patch'.

'

################################################################################
# Validate Inputs

# 'ver_path'
# Must be provided
if [[ -z "${ver_path}" ]]; then
  echo "ERROR: Required input ['ver_path'] is not set."
  exit 1
fi

# 'ver_path'
# Verify the provided VERSION file path exists
if [[ ! -f "${ver_path}" ]]; then
  echo "ERROR: Provided VERSION file path ['${ver_path}'] does not exist."
  # Provide an additional note in CICD
  if [[ -n "${GITHUB_OUTPUT}" ]]; then
    echo "       Did you run actions/checkout first?"
  fi
  exit 1
fi

# 'segment'
# Must be provided
if [[ -z "${segment}" ]]; then
  echo "ERROR: Required input ['segment'] is not set."
  exit 1
fi
segment="${segment,,}" # Lowercase for standard comparison

# 'segment'
# Must be one of: 'major', 'minor' or 'patch'
if ! echo "${segment}" | grep -Po '^(major|minor|patch)$' 2>&1 1>/dev/null; then
  echo "ERROR: Provided input ['segment'] must be one of: 'major', 'minor' or 'patch'."
  exit 1
fi

################################################################################
# Version Bump

# Read the VERSION file
ver="$(cat "${ver_path}")"

# Recognize the major, minor and patch segments.
major="$(echo "${ver}" | grep -Po '^\K[0-9]+')"
minor="$(echo "${ver}" | grep -Po '^[0-9]+\.\K[0-9]+')"
patch="$(echo "${ver}" | grep -Po '^\K[0-9]+\.[0-9]+\.\K[0-9]+')"

# Increment the appropriate version
case "${segment}" in
  "major")
    major=$((major+=1))
    minor=0
    patch=0
    ;;

  "minor")
    minor=$((minor+=1))
    patch=0
    ;;

  "patch")
    patch=$((patch+=1))
    ;;

  *)
    echo "ERROR: Found unexpected version segmnet ['${segment}']."
    exit 1
esac
new_ver="${major}.${minor}.${patch}"

echo "Incrementing version ['${ver}']=>['${new_ver}']."
echo "${new_ver}" > "${ver_path}"
