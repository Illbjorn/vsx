#!/usr/bin/env bash

################################################################################
# Validate inputs

assert_set () {
  if [[ -z "${!1+x}" ]]; then
    echo "ERROR: '${1}' is a required input but was not found."
    exit 1
  fi
}

assert_set 'tag'        # Git tag to create the GitHub release from
assert_set 'title'      # Title to use for the GitHub release
assert_set 'body'       # Body to use for the GitHub release
assert_set 'repo_owner' # GitHub repository owner to create the release in
assert_set 'repo_name'  # GitHub repository name to create the release in
# assert_set 'files'    # Files to attach to the GitHub release

################################################################################
# Create Release

# Disable undeclared variable linting, these values all come from the
# environment.
# shellcheck disable=SC2154
#
# We want globbing with `$files`.
# shellcheck disable=SC2086

# Create release with files attached
if [[ -n "${files}" ]]; then
  gh release create "${tag}" \
    --title "${title}" \
    --notes "${body}" \
    --repo "${repo_owner}/${repo_name}" \
    ${files}
else
  gh release create "${tag}" \
    --title "${title}" \
    --notes "${body}" \
    --repo "${repo_owner}/${repo_name}"
fi
