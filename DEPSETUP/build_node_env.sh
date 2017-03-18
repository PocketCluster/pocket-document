#!/usr/bin/env bash

# This script set up remote host with dependencies for testing, assuming
# 1) SSH password-less connection is setup
# 2) Golang 1.7.5 is supet
# 3) Golang package root path is /opt/gopkg/src

GOREPO=${GOREPO-"${HOME}/Workspace/POCKETPKG"}
WORKDIR=${WORK_ROOT-"${GOREPO}/DEPSETUP"}
TESTGO=${TESTGO:-0}

TARGET_HOST=${1}
if [[ -z ${TARGET_HOST} ]]; then
    echo "Cannot setup null host"
    exit
fi

CONN_CHECK=$(ssh -q -o "BatchMode=yes" ${TARGET_HOST} "echo 2>&1" && echo "OK" || echo "NOPE" )
if [[ ${CONN_CHECK} == "NOPE" ]]; then
    echo "Unable to connect ${TARGET_HOST}"
fi

function rsync_directory() {
    local SRC_DIR=${1}
    local DST_DIR=${2}
    #cd ${SRC_DIR} && rsync -rmvt --no-links -e ssh --exclude '.git*' * "${TARGET_HOST}:${DST_DIR}"
    cd ${SRC_DIR} && rsync -rmvt --no-links -e ssh  --include '*/' --include '*.go' --exclude '*' * "${TARGET_HOST}:${DST_DIR}"
}

function rsync_native_directory() {
    local SRC_DIR=${1}
    local DST_DIR=${2}
    cd ${SRC_DIR} && rsync -rmvt --no-links -e ssh --exclude '.git*' --exclude '.idea' --exclude '*iml' * "${TARGET_HOST}:${DST_DIR}"
}

function clean_remote_files() {
    local DST_FILE=${1}
    ssh "${TARGET_HOST}" "rm -rfv ${DST_FILE}"
}

# check if gobjc++ is installed
# http://stackoverflow.com/questions/8593562/how-to-compile-objc-code-on-linux
# sudo apt-get install gobjc++

pushd "${GOREPO}/DEPSETUP"

echo "make directories..."
ssh "${TARGET_HOST}" 'mkdir -p /opt/gopkg/src/github.com/stkim1'
ssh "${TARGET_HOST}" 'mkdir -p /opt/gopkg/src/github.com/gravitational'

echo "sync dependencies first..."
rsync_native_directory "${GOREPO}/src"                                                "/opt/gopkg/src/"

echo "sync main dependencies..."
rsync_native_directory "${GOREPO}/MAINCOMP/etcd-3.1.1"                                "/opt/gopkg/src/github.com/coreos/etcd/"
rsync_native_directory "${GOREPO}/MAINCOMP/docker-c8388a-2016_11_22"                  "/opt/gopkg/src/github.com/docker/docker/"
rsync_native_directory "${GOREPO}/DEPREPO/teleport"                                   "/opt/gopkg/src/github.com/gravitational/teleport/"

echo "sync dependency projects..."
rsync_native_directory "${HOME}/Workspace/pc-osx-manager/PC-APP-V2/pc-node-agent/"    "/opt/gopkg/src/github.com/stkim1/pc-node-agent/"
rsync_native_directory "${HOME}/Workspace/pc-osx-manager/PC-APP-V2/pc-core/"          "/opt/gopkg/src/github.com/stkim1/pc-core/"
rsync_native_directory "${HOME}/Workspace/pc-osx-manager/GOLANG/pcrypto/"             "/opt/gopkg/src/github.com/stkim1/pcrypto/"
rsync_native_directory "${HOME}/Workspace/pc-osx-manager/GOLANG/netifaces/"           "/opt/gopkg/src/github.com/stkim1/netifaces/"

echo "clean files..."
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/exec/"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/lib/"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/staticlib/"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-node-agent/exec/"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/context/*.h"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/context/*_binding.go"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-core/context/context_darwin.go"
clean_remote_files "/opt/gopkg/src/github.com/stkim1/pc-node-agent/main.go"
clean_remote_files "/opt/gopkg/src/github.com/cloudflare/cfssl/vendor/github.com/jmoiron/sqlx"
ssh ${TARGET_HOST} 'find /opt/gopkg/src/ -name ".DS_Store" | xargs -r rm -v'

popd

if [[ ${TESTGO} -eq 1 ]]; then
    scp "${WORKDIR}/remote_test_dep.sh" "${TARGET_HOST}:~/" && ssh "${TARGET_HOST}" "bash ~/remote_test_dep.sh"
fi
