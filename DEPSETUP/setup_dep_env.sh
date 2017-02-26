#!/usr/bin/env bash

if [[ ! -z ${GOREPO} ]];then
    unset GOREPO
fi
if [[ ! -z ${GOPATH} ]]; then
    unset GOPATH
fi
export GOREPO="${HOME}/Workspace/POCKETPKG"
export GOPATH="$GOREPO:$GOWORKPLACE"
export PATH=$GEM_HOME/ruby/2.0.0/bin:$HOME/.util:$GOROOT/bin:$GOREPO/bin:$GOWORKPLACE/bin:$HOME/.util:$NATIVE_PATH

# workspace root
export WORK_ROOT="${GOREPO}/DEPSETUP"

# main component to unpack
MAIN_COMPONENT=("swarm-1.2.6" "distribution-2.6.0" "etcd-3.1.1" "libcompose-0.4.0")

echo "prep directories"
pushd ${WORK_ROOT}
cd ${GOREPO}
mkdir -p {src,bin,pkg}
popd

${WORK_ROOT}/condense_dep.py && source ${WORK_ROOT}/vendor_cleanup.sh

for comp in ${MAIN_COMPONENT[@]}; do
    if [[ ! -d "${GOREPO}/MAINCOMP/${comp}" ]]; then
        tar -xzf "${GOREPO}/MAINCOMP/${comp}.tar.gz" -C "${GOREPO}/MAINCOMP"
    fi
    # etcd need special attention
    if [[ ${comp} =~ "etcd" ]] && [[ -d "${GOREPO}/MAINCOMP/${comp}/cmd/vendor/" ]]; then
        echo "Special treat for ${comp}..."
        mv "${GOREPO}/MAINCOMP/${comp}/cmd/vendor/" "${GOREPO}/MAINCOMP/${comp}/vendor/"
        rm -rf "${GOREPO}/MAINCOMP/${comp}/cmd/"
    fi
    if [[ -d "${GOREPO}/MAINCOMP/${comp}/vendor" ]]; then
        pushd ${WORK_ROOT}
        echo "cleanup ${GOREPO}/MAINCOMP/${comp}/vendor..."
        cd "${GOREPO}/MAINCOMP/${comp}/vendor" && clean_vendor
        popd
    fi
done

# setup teleport
pushd ${WORK_ROOT}
TELEPORT="${GOREPO}/src/github.com/gravitational/teleport"
mkdir -p ${TELEPORT} && cp -rf ${GOREPO}/TELEPORT/teleport/* "${TELEPORT}/" && cd "${TELEPORT}/vendor/" && clean_vendor
popd

# setup etcd
if [[ -d "${GOREPO}/src/github.com/coreos/etcd" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/coreos/etcd"
    rm "${GOREPO}/src/github.com/coreos/etcd"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/coreos" && cd "${GOREPO}/src/github.com/coreos" && ln -s "../../../MAINCOMP/etcd-3.1.1" "./etcd"
popd

# setup swarm
if [[ -d "${GOREPO}/src/github.com/docker/swarm" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/swarm"
    rm "${GOREPO}/src/github.com/docker/swarm"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/swarm-1.2.6" "./swarm"
popd

# setup libcompose
if [[ -d "${GOREPO}/src/github.com/docker/libcompose" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/libcompose"
    rm "${GOREPO}/src/github.com/docker/libcompose"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/libcompose-0.4.0" "./libcompose"
popd

# setup distribution
if [[ -d "${GOREPO}/src/github.com/docker/distribution" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/distribution"
    rm "${GOREPO}/src/github.com/docker/distribution"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/distribution-2.6.0" "./distribution"
popd
