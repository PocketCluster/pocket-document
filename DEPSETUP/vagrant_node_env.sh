#!/usr/bin/env bash

# This script set up remote host with dependencies for testing, assuming
# 1) This is executed in Vagrant host
# 2) Golang 1.7.5 is used
# 3) Golang package root path is /opt/gopkg/src

GOREPO=${GOREPO-"/Workspace/POCKETPKG"}
WORKDIR=${WORK_ROOT-"${GOREPO}/DEPSETUP"}
TESTGO=${TESTGO:-0}

function rsync_directory() {
    local SRC_DIR=${1}
    local DST_DIR=${2}
    #cd ${SRC_DIR} && rsync -rmvt --no-links -e ssh --exclude '.git*' * "${TARGET_HOST}:${DST_DIR}"
    cd ${SRC_DIR} && rsync -rmvt --no-links --include '*/' --include '*.go' --exclude '*' * "${DST_DIR}"
}

function rsync_native_directory() {
    local SRC_DIR=${1}
    local DST_DIR=${2}
    cd ${SRC_DIR} && rsync -rmvt --no-links --exclude '.git*' --exclude '.idea' --exclude '*iml' * "${DST_DIR}"
}


# check if gobjc++ is installed
# http://stackoverflow.com/questions/8593562/how-to-compile-objc-code-on-linux
# sudo apt-get install gobjc++

pushd "${GOREPO}/DEPSETUP"

echo "clean source root..."
rm -rf /opt/gopkg/src/*

echo "make directories..."
mkdir -p /opt/gopkg/src/github.com/stkim1
mkdir -p /opt/gopkg/src/github.com/gravitational

echo "sync dependencies first..."
rsync_native_directory "${GOREPO}/src"                                         "/opt/gopkg/src/"

echo "sync main dependencies..."
rsync_native_directory "${GOREPO}/MAINCOMP/etcd-3.1.1"                         "/opt/gopkg/src/github.com/coreos/etcd/"
rsync_native_directory "${GOREPO}/MAINCOMP/docker-c8388a-2016_11_22"           "/opt/gopkg/src/github.com/docker/docker/"
rsync_native_directory "${GOREPO}/DEPREPO/teleport"                            "/opt/gopkg/src/github.com/gravitational/teleport/"

echo "sync dependency projects..."
rsync_native_directory "/Workspace/pc-osx-manager/PC-APP-V2/pc-node-agent/"    "/opt/gopkg/src/github.com/stkim1/pc-node-agent/"
rsync_native_directory "/Workspace/pc-osx-manager/PC-APP-V2/pc-core/"          "/opt/gopkg/src/github.com/stkim1/pc-core/"
rsync_native_directory "/Workspace/pc-osx-manager/GOLANG/pcrypto/"             "/opt/gopkg/src/github.com/stkim1/pcrypto/"
rsync_native_directory "/Workspace/pc-osx-manager/GOLANG/netifaces/"           "/opt/gopkg/src/github.com/stkim1/netifaces/"

echo "clean files..."
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/exec/
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/lib/
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/staticlib/
rm -rf /opt/gopkg/src/github.com/stkim1/pc-node-agent/exec/
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/context/*.h
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/context/*_binding.go
rm -rf /opt/gopkg/src/github.com/stkim1/pc-core/context/context_darwin.go
rm -rf /opt/gopkg/src/github.com/stkim1/pc-node-agent/main.go
rm -rf /opt/gopkg/src/github.com/stkim1/netifaces/pysrc/
rm -rf /opt/gopkg/src/github.com/stkim1/netifaces/xcode/
find /opt/gopkg/src/ -name ".DS_Store" | xargs rm

popd

if [[ ${TESTGO} -eq 1 ]]; then
    source "${WORKDIR}/remote_test_dep.sh"
fi
