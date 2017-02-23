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
- Swarm [1.2.6](https://github.com/docker/swarm/releases/tag/v1.2.6)  
  * `go test ./...` passed

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
  git checkout 24a3e3d3fc7451805e09d11e11e95d9a0a4f205e
  ```

