# Assemble Core

## v0.1.4

**Planned**  
- [ ] Check Dependency for Go 1.7.5  
- [ ] Embed ETCD core with garbase cleanup  
- [ ] Embed SWARM with `node://` option, TLS, and authentification  
- [ ] Embed REGISTRY with TLS and authentification  
- [ ] Embed Teleport with BOLTDB  
- [ ] Generate TLS _without_ `IP.1` in `[alt_name]`  
- [ ] Autojoin with DHCP failure incorporated
- [ ] OSX DHCP notification change

**Postponed**  
- [`gomvpkg`](https://godoc.org/golang.org/x/tools/cmd/gomvpkg) is not deployed  

  * <https://gowalker.org/github.com/golang/tools/cmd/gomvpkg>
  * <https://groups.google.com/forum/#!topic/golang-nuts/yBu8lyPmFLM>
  * <https://godoc.org/golang.org/x/tools/cmd/gomvpkg#pkg-files>

**3rd Party Packages**  

All the packages are saved in source code format at <https://github.com/stkim1/pc-osx-manager/PKGS>  

- Golang [1.7.5](https://golang.org/doc/go1.7)
  * [Compiled and passed for ARMHF, ARM64](https://github.com/stkim1/GOLANG-ARM)  
- ETCD [3.1.1](https://github.com/coreos/etcd/releases/tag/v2.3.8)  
  * Recommended to use with Golang 1.7+
  * use `github.com/coreos/etcd/embed`
- Swarm [1.2.6](https://github.com/docker/swarm/releases/tag/v1.2.6)  
  * `go test ./...` passed
  * `$GOPATH/src/github.com/docker/swarm && go install ./...` will install `swarm` binary in `$GOPATH/bin`
- Distribution [2.6.0](https://github.com/docker/distribution/releases/tag/v2.6.0)
  * `cd $GOPATH/src/github.com/docker/distribution/cmd/registry && go install` will install `registry` binary in `$GOPATH/bin`
- Libcompose [0.4.0](https://github.com/docker/libcompose/releases/tag/v0.4.0) + [42066b](https://github.com/docker/libcompose/commit/42066b8afe5c987ac896076f704a87a35cee86c5) + [f5739a](https://github.com/docker/libcompose/commit/f5739a73c53493ebd1ff76d6ec95f3fc1c478c38)
- Teleport [1.2.0-baafe3](https://github.com/gravitational/teleport/commit/baafe3a332735d0cf7111be8ad571869fe038b35)
  
  ```sh
  # Clone single branch from origin
  git clone -b master --single-branch https://github.com/gravitational/teleport
  
  # Add `private` as a mirroring remote
  git remote add private https://github.com/stkim1/teleport.git
  
  # check remotes
  git remote -v
  
  origin  https://github.com/gravitational/teleport (fetch)
  origin  https://github.com/gravitational/teleport (push)
  private https://github.com/stkim1/teleport.git (fetch)
  private https://github.com/stkim1/teleport.git (push)
  
  # fetch origin master
  git pull origin
  
  # update `private` master
  git push private master
  
  # create backbone branch from baafe3a332735d0cf7111be8ad571869fe038b35
  git checkout -b backbone baafe3a332735d0cf7111be8ad571869fe038b35
  
  # setup the branch
  git push --set-upstream private backbone
  ```

  > Reference
  
  - [Teleport 07e8d2...33044f](https://github.com/gravitational/teleport/compare/07e8d212ff5caff194feb4b217b4638e238d0c86...33044f6d89525e3055a33620b5877db1320576a5)
  - [Teleport 1d0ec4](https://github.com/gravitational/teleport/commit/1d0ec48dfa788d03f016a6754219dd67db890c8f)
  - [Teleport baafe3](https://github.com/gravitational/teleport/commit/baafe3a332735d0cf7111be8ad571869fe038b35)


## Issues

### `ETCD`

- [etcd < v3.1 does not work properly if built with Go > v1.7](https://github.com/coreos/etcd/blob/master/Documentation/upgrades/upgrade_3_0.md#known-issues). Issue [6951](https://github.com/coreos/etcd/issues/6951)
  * Use ETCD 3.1.1 with Golang 1.7.5

### `Swarm`
- Swarm 1.2.6 has platform incompatibility issue with [Microsoft/go-winio](https://github.com/Microsoft/go-winio). Issue [swarmkit#1067](https://github.com/docker/swarmkit/issues/1067)<br/> The issue is patched at [d8f60f2](https://github.com/Microsoft/go-winio/commit/d8f60f2dd117cd64c2825143a89ecb6f158ad743) and Go 1.7 compatibility is checked at [24a3e3](https://github.com/Microsoft/go-winio/commit/24a3e3d3fc7451805e09d11e11e95d9a0a4f205e)
  * Patch Microsoft/go-winio in `swarm/Godeps/Godeps.json`

  ```json
  {
    "ImportPath": "github.com/Microsoft/go-winio",
  - "Rev": "c40bf24f405ab3cc8e1383542d474e813332de6d"
  + "Rev": "24a3e3d3fc7451805e09d11e11e95d9a0a4f205e"
  },
  ```

  ```sh
  cd vendor/github.com/Microsoft
  rm -rf go-winio
  git clone https://github.com/Microsoft/go-winio
  cd go-winio
  git checkout 24a3e3d3fc7451805e09d11e11e95d9a0a4f205e
  ```

### `Distribution` (Registry)  

- `github.com/docker/distribution/registry/storage/driver/inmemory` test failed due to prolonged test time


### `Libcompose`  

- `github.com/docker/libcompose-0.4.0` collide with docker 
  * The latest version `f5739a` refers a newest docker version that would not be compatible with what `swarm` use (`c8388a`).
  * In order to mitigate the difference, `0.4.0` is used as base and `42066b` (disable oom killer) + `f5739a` (new docker engine incorporation).
  * We don't need following storage dependencies. So removed are they and their vendors
    - s3      :   github.com/aws/aws-sdk-go, github.com/docker/goamz
    - azure   :   github.com/Azure/azure-sdk-for-go (collided with docker)
    - gcs     :   google.golang.org/cloud (collided with docker)
    - swift   :   github.com/ncw/swift
    - oss     :   github.com/denverdino/aliyungo

    ```sh
    google.golang.org/cloud
    ----------------- !!!CONFLICT!!! --------------------- 
    dae7e3d993bc3812a2185af60552bb6b847e52a0    ->    2015-12-16 16:54:51+11:00 : docker-c8388a-2016_11_22
    975617b05ea8a58727e6c1a06b6161ff4185a9f2    ->    2015-11-04 17:14:34-05:00 : distribution-2.6.0
    
    
    github.com/aws/aws-sdk-go
    ----------------- !!!CONFLICT!!! --------------------- 
    v1.4.22                                     ->    2016-10-25 21:23:00+00:00 : docker-c8388a-2016_11_22
    90dec2183a5f5458ee79cbaf4b8e9ab910bc81a6    ->    2016-07-07 17:08:20-07:00 : distribution-2.6.0
	```

    ```sh
    rm -rf ./distribution/registry/storage/driver/azure/
    rm -rf ./distribution/registry/storage/driver/gcs/
    rm -rf ./distribution/registry/storage/driver/oss/
    rm -rf ./distribution/registry/storage/driver/s3-aws/
    rm -rf ./distribution/registry/storage/driver/s3-goamz/
    rm -rf ./distribution/registry/storage/driver/swift/
    rm -rf ./distribution/registry/storage/driver/middleware/cloudfront/

    rm -rf ./distribution/vendor/github.com/aws/aws-sdk-go/
    rm -rf ./distribution/vendor/github.com/docker/goamz/
    rm -rf ./distribution/vendor/github.com/Azure/azure-sdk-for-go/
    rm -rf ./distribution/vendor/google.golang.org/cloud/
    rm -rf ./distribution/vendor/github.com/ncw/swift/
    rm -rf ./distribution/vendor/github.com/denverdino/aliyungo/
    ```
  * `github.com/Sirupsen/logrus` (f76d643702a30fbffecdfe50831e11881c96ceb3) : logrus is aligned with logrus in docker-c8388a-2016_11_22
  * `github.com/bshuster-repo/logrus-logstash-hook` (5f729f2fb50a301153cae84ff5c58981d51c095a) : Formatter version is aligned with [distribution 50133d] (https://github.com/docker/distribution/blob/50133d63723f8fa376e632a853739990a133be16/vendor.conf)
  * `github.com/cpuguy83/go-md2man` (a65d4d2de4d5f7c74868dfa9b202a3c8be315aaa) : This is added due to original godep misses. Aligned with docker-c8388a-2016_11_22

### `Teleport`  

- Why go with old version?
  * We don't want "cluster snapshot" that works without `auth`
  * We don't need 2FA hardware support
  * We need to work with legacy codebase that has been modified
  * `baafe3` commit will be a base of [`backbone`](https://github.com/stkim1/teleport/tree/backbone) branch and we'll make modifications to that with cherry picking.
