#!/usr/bin/env bash

source ./util.sh

POCKET_REPO="${HOME}/.pocket/repository/"

pushd ${PWD}
cd ${POCKET_REPO} && rm -rf * && mkdir -p docker/registry/v2/repositories && mkdir -p docker/registry/v2/blobs
popd
