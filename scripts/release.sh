#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN is not set. Export it before running this script."
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree is not clean. Commit or stash changes before releasing."
  git status --short
  exit 1
fi

echo "Enter release version (e.g., 1.1.1):"
read -r VERSION_INPUT

if [[ -z "${VERSION_INPUT}" ]]; then
  echo "Version is required."
  exit 1
fi

if [[ "${VERSION_INPUT}" == v* ]]; then
  VERSION="${VERSION_INPUT}"
else
  VERSION="v${VERSION_INPUT}"
fi

echo "Releasing ${VERSION}"

git tag -a "${VERSION}" -m "Release ${VERSION}"

git push origin main

git push origin "${VERSION}"

GORELEASER_CURRENT_TAG="${VERSION}" goreleaser release --clean

echo "Release complete: ${VERSION}"
