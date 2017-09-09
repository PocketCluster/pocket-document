#!/usr/bin/env bash

BLOCKSIZE=$((1024*1024*8))
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

echo "BlockSize ${BLOCKSIZE}, Core image : ${COREIMAGE}, Node image : ${NODEIMAGE}, Meta Json : ${METAJSON}"

CORE_CHKSUM=$(pcsync build --blocksize ${BLOCKSIZE} --quite ${COREIMAGE})
NODE_CHKSUM=$(pcsync build --blocksize ${BLOCKSIZE} --quite ${NODEIMAGE})
META_CHKSUM=$(pcsync meta ${METAJSON})
PKG_VER=$(pcsync pkgver ${CORE_CHKSUM} ${NODE_CHKSUM} ${META_CHKSUM})

pushd ${PWD}
echo "syncing formula..."
cd ../v014 && find . -name ".DS_Store" -depth -exec rm {} \; && rsync -rlmv ./* api:/api-service/v014
popd
