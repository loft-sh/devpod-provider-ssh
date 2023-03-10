#!/usr/bin/env bash
set -e

PROVIDER_ROOT=$(git rev-parse --show-toplevel)

echo "Current working directory is $(pwd)"
echo "PATH is $PATH"

if [[ "$(pwd)" != "${PROVIDER_ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${PROVIDER_ROOT} before running this script" 1>&2
  exit 1
fi
mkdir -p "${PROVIDER_ROOT}/release"

# generate provider.yaml
 cat hack/provider/provider.yaml | sed "s/##VERSION##/${RELEASE_VERSION}/g" > "${PROVIDER_ROOT}/release/provider.yaml"
