#!/usr/bin/env bash

source ./util.sh

DEST_ARCH=${1}
POCKET_REPO="${HOME}/.pocket/repository/"

if [[ -z ${DEST_ARCH} ]]; then
  echo "please specify destination archive name w/ tar.xz"
fi

pushd ${PWD}
cd ../archive
DEST_ARCH="$(pwd)/${DEST_ARCH}.tar.xz"
popd

echo ${DEST_ARCH}
rm ${DEST_ARCH}

pushd ${PWD}
cd ${POCKET_REPO}
tar cvf - . | xz -zf - > ${DEST_ARCH}
popd
