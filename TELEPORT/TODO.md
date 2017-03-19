# TODO on integration

## v0.1.4 Release Targets

- [ ] Generate HostUUID from config and use it across application
- [ ] Dynamically invite node to join
- [ ] Check **SCP** is working
- [ ] `lib/auth/pc_local_auth.go:RequestHOTPforSignupToken()` arguments needs encryption
- [ ] Node health Check
- [ ] Adjust heart-beat frequency
- [ ] **it should be possible to issue a cert without IP Address** -> DHCP related issue
  - Implement own DNS network
- [ ] Check what happens when there is no auth server on net and no certificate exists on local machine.
  - Make it less severe
- [ ] Check what happens when IP address change. Make sure if it's necessary to refresh ARP table
- [ ] Remove directory creation
- [ ] Disable, not remove audit for release
- [ ] See if AES encryption is necessary for `lib/auth/pc_tun/AuthAESEncryption`
- [ ] Add action perm for requesting signed certificate to `lib/auth/pc_auth_with_roles/issueSignedCertificateWithToken`
- [ ] Check if signed cert for this uuid exists. If does, return the value `lib/auth/pc_auth_with_roles/createSignedCertificate`


## Next Phase

- [ ] Integrate CFSSL and have secure, hierarchical cert management.
  - We needs temporary certs for container, multi-root signing, etc.
- [ ] Reduce access points in `APIServer`, `AuthWithRoles`.
- [ ] Deliver parameters down to `lib/auth/native/nauth` and incorporate with CFSSL.
- [ ] `lib/auth/pc_local_auth.go:RequestHOTPforSignupToken()` needs to be removed with Identity service.
- [ ] Remove entire `lib/web` package
- [ ] Use `keyAuth` only in `APIServer` to auth node, user instead of password
- [ ] Use Identity service on OSX
- [ ] Add known hosts and persist it with heart-beat
