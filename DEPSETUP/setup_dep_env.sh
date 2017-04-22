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
MAIN_COMPONENT=("swarm-1.2.6" "distribution-2.6.0" "docker-c8388a-2016_11_22")

# TESTING
TESTGO=${TESTGO:-0}
ADV_TESTGO=${ADV_TESTGO:-0}
# COPYING REPO
COPY_DEP_REPO=${COPY_DEP_REPO:-0}

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

if [[ ${COPY_DEP_REPO} -eq 1 ]]; then

    # setup teleport
    pushd ${WORK_ROOT}
    echo "setting up teleport..."
    TELEPORT="${GOREPO}/src/github.com/gravitational/teleport"
    if [[ -d ${TELEPORT} ]]; then
        #find ${TELEPORT} -mindepth 1 -maxdepth 1 ! -name '*.iml' | xargs rm -rf
        rm -rf ${TELEPORT}/*
    else
        mkdir -p ${TELEPORT}
    fi
    cp -rf ${GOREPO}/DEPREPO/teleport/* ${TELEPORT}/ && cd "${TELEPORT}/vendor/" && clean_vendor && (rm ${TELEPORT}/*.iml || true)
    popd

    # setup libcompose
    pushd ${WORK_ROOT}
    echo "setting up libcompose..."
    LIBCOMPOSE="${GOREPO}/src/github.com/docker/libcompose"
    if [[ -d ${LIBCOMPOSE} ]]; then
        #find ${LIBCOMPOSE} -mindepth 1 -maxdepth 1 ! -name '*.iml' | xargs rm -rf
        rm -rf ${LIBCOMPOSE}/*
    else
        mkdir -p ${LIBCOMPOSE}
    fi
    cp -rf ${GOREPO}/DEPREPO/libcompose/* ${LIBCOMPOSE}/ && cd "${LIBCOMPOSE}/vendor/" && clean_vendor && (rm ${LIBCOMPOSE}/*.iml || true)
    popd

    # setup etcd
    pushd ${WORK_ROOT}
    echo "setting up etcd..."
    ETCD="${GOREPO}/src/github.com/coreos/etcd"
    if [[ -d ${ETCD} ]]; then
        rm -rf ${ETCD}/*
    else
        mkdir -p ${ETCD}
    fi
    cp -rf ${GOREPO}/DEPREPO/etcd/* ${ETCD}/ && cd "${ETCD}/vendor/" && clean_vendor && (rm ${ETCD}/*.iml || true)
    popd

else

    # setup teleport
    pushd ${WORK_ROOT}
    echo "setting up teleport..."
    GRAVITATIONAL="${GOREPO}/src/github.com/gravitational"
    if [[ ! -d ${GRAVITATIONAL} ]]; then
        mkdir -p "${GRAVITATIONAL}/"
    fi
    TELEPORT="${GRAVITATIONAL}/teleport"
    LINK=$(readlink "${TELEPORT}")
    if [[ ! -d ${TELEPORT} ]] || [[ $LINK != "../../../DEPREPO/teleport" ]]; then
        echo "cleanup old link ${TELEPORT} and rebuild..."
        cd "${GRAVITATIONAL}/" && (rm ${TELEPORT} || true) && ln -s ../../../DEPREPO/teleport ./teleport
    fi
    cd "${TELEPORT}/vendor/" && clean_vendor
    popd

    # setup libcompose
    pushd ${WORK_ROOT}
    echo "setting up libcompose..."
    DOCKER="${GOREPO}/src/github.com/docker"
    if [[ ! -d ${DOCKER} ]]; then
        mkdir -p "${DOCKER}/"
    fi
    LIBCOMPOSE="${DOCKER}/libcompose"
    LINK=$(readlink "${LIBCOMPOSE}")
    if [[ ! -d ${LIBCOMPOSE} ]] || [[ $LINK != "../../../DEPREPO/libcompose" ]]; then
        echo "cleanup old link ${LIBCOMPOSE} and rebuild..."
        cd ${DOCKER} && (rm ${LIBCOMPOSE} || true) && ln -s ../../../DEPREPO/libcompose ./libcompose
    fi
    cd "${LIBCOMPOSE}/vendor/" && clean_vendor
    popd

    pushd ${WORK_ROOT}
    echo "setting up etcd..."
    COREOS="${GOREPO}/src/github.com/coreos"
    if [[ ! -d ${COREOS} ]]; then
        mkdir -p "${COREOS}"
    fi
    ETCD="${COREOS}/etcd"
    LINK=$(readlink "${ETCD}")
    if [[ ! -d ${ETCD} ]] || [[ ${LINK} != "../../../DEPREPO/etcd" ]]; then
        echo "cleanup old link ${ETCD} and rebuild..."
        cd ${COREOS} && (rm ${ETCD} || true) && ln -s ../../../DEPREPO/etcd ./etcd
    fi
    if [[ -d "${ETCD}/vendor/" ]]; then
        cd "${ETCD}/vendor/" && clean_vendor
    else
        #cd ${ETCD} && ${GOREPO}/bin/glide install
        echo "!!! SETUP ETCD VENDOR WITH GLIDE FIRST !!!"
    fi
    popd

fi

# setup swarm
pushd ${WORK_ROOT}
echo "setting up swarm..."
if [[ -d "${GOREPO}/src/github.com/docker/swarm" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/swarm"
    rm "${GOREPO}/src/github.com/docker/swarm"
else
    mkdir -p "${GOREPO}/src/github.com/docker/"
fi
cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/swarm-1.2.6" "./swarm"
popd

# setup distribution
pushd ${WORK_ROOT}
echo "setting up distribution..."
if [[ -d "${GOREPO}/src/github.com/docker/distribution" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/distribution"
    rm "${GOREPO}/src/github.com/docker/distribution"
else
    mkdir -p "${GOREPO}/src/github.com/docker/"
fi
cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/distribution-2.6.0" "./distribution"
echo "special treatment for distribution/registry/registry.go since registry.go is using very old version of logrus, it still contains logstash formatter"
sed -i '' 's|"github.com/Sirupsen/logrus/formatters/logstash"|logstash "github.com/bshuster-repo/logrus-logstash-hook"|g' ./distribution//registry/registry.go

echo "eliminate storage driver other than filesystem..."
rm -rf ./distribution/registry/storage/driver/azure/
rm -rf ./distribution/registry/storage/driver/gcs/
rm -rf ./distribution/registry/storage/driver/oss/
rm -rf ./distribution/registry/storage/driver/s3-aws/
rm -rf ./distribution/registry/storage/driver/s3-goamz/
rm -rf ./distribution/registry/storage/driver/swift/
rm -rf ./distribution/registry/storage/driver/middleware/cloudfront/

echo "eliminate vendored storage drivers..."
rm -rf ./distribution/vendor/github.com/aws/aws-sdk-go/           && (rmdir ./distribution/vendor/github.com/aws > /dev/null 2>&1 || true)
rm -rf ./distribution/vendor/github.com/docker/goamz/             && (rmdir ./distribution/vendor/github.com/docker > /dev/null 2>&1 || true)
rm -rf ./distribution/vendor/github.com/Azure/azure-sdk-for-go/   && (rmdir ./distribution/vendor/github.com/Azure > /dev/null 2>&1 || true)
rm -rf ./distribution/vendor/google.golang.org/cloud/             && (rmdir ./distribution/vendor/google.golang.org > /dev/null 2>&1 || true)
rm -rf ./distribution/vendor/github.com/ncw/swift/                && (rmdir ./distribution/vendor/github.com/ncw > /dev/null 2>&1 || true)
rm -rf ./distribution/vendor/github.com/denverdino/aliyungo/      && (rmdir ./distribution/vendor/github.com/denverdino > /dev/null 2>&1 || true)

popd

# setup docker
if [[ -d "${GOREPO}/src/github.com/docker/docker" ]]; then
    echo "delete old link : ${GOREPO}/src/github.com/docker/docker"
    rm "${GOREPO}/src/github.com/docker/docker"
fi
pushd ${WORK_ROOT}
mkdir -p "${GOREPO}/src/github.com/docker/" && cd "${GOREPO}/src/github.com/docker" && ln -s "../../../MAINCOMP/docker-c8388a-2016_11_22" "./docker"
popd

# THIS IS 2ndary dependency setup. We might need to recover from this!
# cfssl dependency clearing
if [[ ! -d "${GOREPO}/src/github.com/cloudflare/cfssl" ]]; then
    echo "Cloudflare cfssl is not present!!!"
else
    # (03/18/17)
    # with combined dependency clearing, we get "WebSuite.TestNodesWithSessions" error in
    # "github.com/gravitational/teleport/lib/web" testing. Just remove sqlx for now.

    echo "[2ND DEP] Clearing Cloudflare cfssl dependencies..."
    pushd ${WORK_ROOT}
    cd "${GOREPO}/src/github.com/cloudflare/cfssl" && git stash
    rm -rf "${GOREPO}/src/github.com/cloudflare/cfssl/vendor/github.com/jmoiron/sqlx"
    popd
fi

if [[ ! -d "${GOREPO}/src/github.com/getsentry/raven-go" ]]; then
    echo "Senty.io raven-go is not present!!!"
    exit 1
else
    echo "[2ND DEP] Setup raven-go dep"
    pushd ${WORK_ROOT}
    if [[ ! -d "${GOREPO}/src/github.com/getsentry/raven-go/vendor/github.com/certifi/gocertifi" ]]; then
        mkdir -p "${GOREPO}/src/github.com/getsentry/raven-go/vendor/github.com/certifi"
        cd "${GOREPO}/src/github.com/getsentry/raven-go/vendor/github.com/certifi"
        git clone https://github.com/certifi/gocertifi
    fi
    cd "${GOREPO}/src/github.com/getsentry/raven-go/vendor/github.com/certifi/gocertifi"
    git checkout 03be5e6bb9874570ea7fb0961225d193cbc374c5
    popd
fi

if [[ ${TESTGO} -eq 1 ]]; then
source ${WORK_ROOT}/test_dep.sh
fi
