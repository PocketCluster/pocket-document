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
- Libcompose [0.4.0 -> 70abeb](https://github.com/docker/libcompose/commit/70abeb3a42fce124d6b688e39a7c2153f3a363ad) + [f5739a](https://github.com/docker/libcompose/commit/f5739a73c53493ebd1ff76d6ec95f3fc1c478c38)  

  ```sh
  # Clone single branch from origin
  git clone -b master --single-branch https://github.com/docker/libcompose
  
  # Add `private` as a mirroring remote
  git remote add private git@github.com:stkim1/libcompose.git
  
  # check remotes
  git remote -v
  
  origin  https://github.com/gravitational/teleport (fetch)
  origin  https://github.com/gravitational/teleport (push)
  private	git@github.com:stkim1/libcompose.git (fetch)
  private	git@github.com:stkim1/libcompose.git (push)
  
  # fetch origin master
  git pull origin master
  
  # update `private` master
  git push private master
  
  # create backbone branch from 70abeb3a42fce124d6b688e39a7c2153f3a363ad
  git checkout -b backbone 70abeb3a42fce124d6b688e39a7c2153f3a363ad
  
  # setup the branch
  git push --set-upstream private backbone
  ```

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
  - <http://stackoverflow.com/questions/7167645/how-do-i-create-a-new-git-branch-from-an-old-commit>

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
- We don't need following storage dependencies. So removed are they and their vendors
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
- `github.com/Sirupsen/logrus` (f76d643702a30fbffecdfe50831e11881c96ceb3) : logrus is aligned with logrus in docker-c8388a-2016_11_22
- `github.com/bshuster-repo/logrus-logstash-hook` (5f729f2fb50a301153cae84ff5c58981d51c095a) : Formatter version is aligned with [distribution 50133d] (https://github.com/docker/distribution/blob/50133d63723f8fa376e632a853739990a133be16/vendor.conf)
- `github.com/cpuguy83/go-md2man` (a65d4d2de4d5f7c74868dfa9b202a3c8be315aaa) : This is added due to original godep misses. Aligned with docker-c8388a-2016_11_22

### `Libcompose`  

- `github.com/docker/libcompose-0.4.0` is incompatible with docker-c8388a-2016_11_22. 
  * The latest version `f5739a` refers newer docker version than `c8388a` that would not be compatible with what `swarm` use (`c8388a`).
  * In order to mitigate the difference, `0.4.0` is used as base and `70abeb` (null volume configurations) + `f5739a` (new docker engine incorporation).

### `Teleport`  

- Why go with old version?
  * We don't want "cluster snapshot" that works without `auth`
  * We don't need 2FA hardware support
  * We need to work with legacy codebase that has been modified
  * `baafe3` commit will be a base of [`backbone`](https://github.com/stkim1/teleport/tree/backbone) branch and we'll make modifications to that with cherry picking.

### `Docker`

- Docker [1.10.3](https://github.com/docker/docker/releases/tag/v1.10.3), [commit 20f81d](https://github.com/docker/docker/commit/20f81dde9bd97c86b2d0e33bbbf1388018611929)
- Docker [1.11.0](https://github.com/docker/docker/releases/tag/v1.11.0), [commit 4dc599](https://github.com/docker/docker/commit/4dc5990d7565a4a15d641bc6a0bc50a02cfcf302)
- Difference [20f81d...4dc599](https://github.com/docker/docker/compare/20f81dde9bd97c86b2d0e33bbbf1388018611929...4dc5990d7565a4a15d641bc6a0bc50a02cfcf302)

### [golang.org/x/crypto](https://golang.org/x/crypto)

ARMv7 + ARM64 don't pass tests with old version so upgrade to commit [728b75](https://github.com/golang/crypto/commit/728b753d0135da6801d45a38e6f43ff55779c5c2) (Jan 24, 2017)<br/>
<sup>*</sup>[3fbbcd](https://github.com/golang/crypto/commit/3fbbcd23f1cb824e69491a5930cfeff09b12f4d2) for `docker-c8388a-2016_11_22` (2016-04-06) is discarded

### Git : `fatal: bad object`

When this happens, we cannot determine the commit date which lead to inaccurate dependency setup.

In order to fix this, we can use github api and pull data but it's combersome.
Look out if `2000-01-01 00:00:00+00:00` happens on a package, check the packages github page to determine the date again.
Plus, do not use `--branch master` flag for github as it occationally skips commit data.

Current package with bad object and accurate dependency date.

- <https://golang.org/x/crypto>
  * `[728b75](https://github.com/golang/crypto/commit/728b753d0135da6801d45a38e6f43ff55779c5c2) -> 2017-01-23 17:46:54-08:00 : pocketcluster-req-0.1.4`
  * Re-download and check if commit shows correct date
- <https://github.com/spf13/cobra>
  * `[v1.5](https://github.com/dnephin/cobra/commit/0e9ca70a23585bdaab4beba1e7c4c23a1adfa857) -> 2016-11-03 18:13:39-04:00 : docker-c8388a-2016_11_22`
  * download a folk from <https://github.com/dnephin/cobra> and checkout again
- <https://golang.org/x/net>
  * `[f24994](https://github.com/golang/net/commit/f2499483f923065a842d38eb4c7f1927e6fc6e6d) -> 2017-01-14 15:22:49+11:00 : etcd-3.1.1`
  * check the commit date from github