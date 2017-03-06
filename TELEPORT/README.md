# Teleport Port

- No U2FS, Zipped assets, RBAC, and Multi-cluster will be supported.

- Teleport version from [commit bed42f](https://github.com/gravitational/teleport/commit/bed42f3c89111bf75b32aca53434d724e9cd48d0) gets complicated with multi-cluster support and [Role-Based-Acess-Control](https://github.com/gravitational/teleport/issues/620) [PR PT #1](https://github.com/gravitational/teleport/pull/652), which introduces unnecessary complication.

- We can generate custom user cert extension to store metadata <https://github.com/gravitational/teleport/issues/620>

  > It is worth noting that we can't use custom user cert extensions to store metadata to keep its certificates compatible with OpenSSH due to [this bug](https://bugzilla.mindrot.org/show_bug.cgi?id=2387)
