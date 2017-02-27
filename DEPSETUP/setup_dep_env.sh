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
MAIN_COMPONENT=("swarm-1.2.6" "distribution-2.6.0" "etcd-3.1.1" "libcompose-0.4.0-f5739a7-mix" "docker-c8388a-2016_11_22")

# TESTING
TESTGO=${TESTGO:-1}
ADV_TESTGO=${ADV_TESTGO:-0}

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
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/libcompose-0.4.0-f5739a7-mix" "./libcompose"
popd

# setup distribution
if [[ -d "${GOREPO}/src/github.com/docker/distribution" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/distribution"
    rm "${GOREPO}/src/github.com/docker/distribution"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/distribution-2.6.0" "./distribution"
echo "special treatment for distribution/registry/registry.go since registry.go is using very old version of logrus, it still contains logstash formatter"
sed -i '' 's|"github.com/Sirupsen/logrus/formatters/logstash"|logstash "github.com/bshuster-repo/logrus-logstash-hook"|g' ./distribution/registry/registry.go
popd

# setup docker
if [[ -d "${GOREPO}/src/github.com/docker/docker" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/docker"
    rm "${GOREPO}/src/github.com/docker/docker"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/docker-c8388a-2016_11_22" "./docker"
popd



if [[ ${TESTGO} -eq 1 ]]; then
pushd ${WORK_ROOT}
echo "Starting testing..."

# Testing teleport
:<<STATIC_HOST_JOIN_FAIL
* This fails at the first attempt, but succeed in the subsequent one
----------------------------------------------------------------------
FAIL: tun_test.go:425: TunSuite.TestSync

tun_test.go:461:
    c.Assert(syncedServers, DeepEquals, discoveredServers)
... obtained []utils.NetAddr = []utils.NetAddr(nil)
... expected []utils.NetAddr = []utils.NetAddr{utils.NetAddr{Addr:"node.example.com:12345", AddrNetwork:"tcp", Path:""}}

OOPS: 20 passed, 1 FAILED
--- FAIL: TestAPI (5.94s)
STATIC_HOST_JOIN_FAIL
cd "${GOREPO}/src/github.com/gravitational/teleport/lib"            && go test ./...


# Testing etcd
if [[ ${ADV_TESTGO} -eq 1 ]];then
# it passes but it takes long, 266.584s
    cd "${GOREPO}/src/github.com/coreos/etcd/integration/"          && go test ./...
fi
# !!! some tests in v3 client does not pass  !!!
# cd "${GOREPO}/src/github.com/coreos/etcd/clientv3/"               && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/auth/"                     && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/client/"                   && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/compactor/"                && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/contrib/"                  && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/discovery/"                && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/embed/"                    && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/error/"                    && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/etcdserver/"               && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/lease/"                    && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/mvcc/"                     && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/pkg/"                      && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/proxy/"                    && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/raft/"                     && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/rafthttp/"                 && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/snap/"                     && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/store/"                    && go test ./...
cd "${GOREPO}/src/github.com/coreos/etcd/wal/"                      && go test ./...


# Testing swarm
cd "${GOREPO}/src/github.com/docker/swarm/cluster/"                 && go test ./...
cd "${GOREPO}/src/github.com/docker/swarm/api/"                     && go test ./...
cd "${GOREPO}/src/github.com/docker/swarm/discovery/"               && go test ./...
cd "${GOREPO}/src/github.com/docker/swarm/scheduler/"               && go test ./...


# Testing libcompose
if [[ ${ADV_TESTGO} -eq 1 ]];then
# we need an actual docker daemon running for integration test
    cd "${GOREPO}/src/github.com/docker/libcompose/integration/"    && go test ./...
fi
cd "${GOREPO}/src/github.com/docker/libcompose/test/"               && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/docker/"             && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/config/"             && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/labels/"             && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/logger/"             && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/lookup/"             && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/project/"            && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/utils/"              && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/version/"            && go test ./...
cd "${GOREPO}/src/github.com/docker/libcompose/yaml/"               && go test ./...


# Testing distribution
if [[ ${ADV_TESTGO} -eq 1 ]];then
# github.com/docker/distribution/registry/storage/driver/inmemory takes too long to test
    cd "${GOREPO}/src/github.com/docker/distribution/registry"      && go test ./...
fi
cd "${GOREPO}/src/github.com/docker/distribution/configuration"     && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/context"           && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/digest"            && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/health"            && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/manifest"          && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/notifications"     && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/reference"         && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/testutil"          && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/uuid"              && go test ./...
cd "${GOREPO}/src/github.com/docker/distribution/version"           && go test ./...

echo "Testing Done!"
popd
fi