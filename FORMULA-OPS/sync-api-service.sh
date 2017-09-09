#!/usr/bin/env bash

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd "$@" > /dev/null
}

BLOCKSIZE=$((1024 * 1024 * 8))
COREIMAGE=${1}
NODEIMAGE=${2}
METAJSON=${3}

if [[ ${PWD} != *"pocket-formula/devops" ]]; then
  echo "should be executed in pocket-formula/devops directory"
  exit
fi

if [[ -z ${COREIMAGE} ]]; then
  echo "should specify core image"
  exit
fi

if [[ -z ${NODEIMAGE} ]]; then
  echo "should specify node image"
  exit
fi

if [[ -z ${METAJSON} ]]; then
  echo "should specify meta json path"
  exit
fi

CORE_CHKSUM=$(pcsync build --blocksize ${BLOCKSIZE} --quite --output-dir ../v014/package/sync/ ${COREIMAGE})
NODE_CHKSUM=$(pcsync build --blocksize ${BLOCKSIZE} --quite --output-dir ../v014/package/sync/ ${NODEIMAGE})
META_CHKSUM=$(pcsync meta ${METAJSON})
PKG_VER=$(pcsync pkgver ${CORE_CHKSUM} ${NODE_CHKSUM} ${META_CHKSUM})

printf "BlockSize ${BLOCKSIZE}\n\tCore image : ${COREIMAGE} -> ${CORE_CHKSUM}\n\tNode image : ${NODEIMAGE} -> ${NODE_CHKSUM}\n\tMeta Json  : ${METAJSON} -> ${META_CHKSUM}\n"

pcsync pkglist ${CORE_CHKSUM} ${NODE_CHKSUM} ${META_CHKSUM} ${PKG_VER} ../template/base-package.template ../v014/package/list.json

pushd ${PWD} > /dev/null
echo "syncing formula..."
cd ../v014 && find . -name ".DS_Store" -depth -exec rm {} \; && rsync -rlmv ./* api:/api-service/v014
popd
