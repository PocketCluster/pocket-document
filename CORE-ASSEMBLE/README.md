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
- Libcompose [0.4.0](https://github.com/docker/libcompose/releases/tag/v0.4.0)
- Teleport [1.2.0-baafe3](https://github.com/gravitational/teleport/commit/baafe3a332735d0cf7111be8ad571869fe038b35)
  * We don't want "cluster snapshot" that works without `auth`
  * We don't need 2FA hardware support
  * We need to work with legacy codebase that has been modified
  * `baafe3` commit will be a base of [`backbone`](https://github.com/stkim1/teleport/tree/backbone) branch and we'll make modifications to that with cherry picking.
  
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


**Issues**

- [etcd < v3.1 does not work properly if built with Go > v1.7](https://github.com/coreos/etcd/blob/master/Documentation/upgrades/upgrade_3_0.md#known-issues). Issue [6951](https://github.com/coreos/etcd/issues/6951)
  * Use ETCD 3.1.1 with Golang 1.7.5
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
- `github.com/docker/distribution/registry/storage/driver/inmemory` test failed due to prolonged test time
