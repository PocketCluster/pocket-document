
/// --- lib/auth/tun.go --- ///
// NewTunnel creates a new SSH tunnel server which is not started yet.
// This is how "site API" (aka "auth API") is served: by creating
// an "tunnel server" which serves HTTP via SSH.
func NewTunnel(addr utils.NetAddr,
	hostSigner ssh.Signer,
	apiConf *APIConfig,
	opts ...ServerOption) (tunnel *AuthTunnel, err error) {

	  = &AuthTunnel{
		authServer: apiConf.AuthServer,
		config:     apiConf,
	}
	tunnel.limiter, err = limiter.NewLimiter(limiter.LimiterConfig{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	// apply functional options:
	for _, o := range opts {
		if err := o(tunnel); err != nil {
			return nil, err
		}
	}
	// create an SSH server and assign the tunnel to be it's "new SSH channel handler"
	tunnel.sshServer, err = sshutils.NewServer(
		teleport.ComponentAuth,
		addr,
		tunnel,
		[]ssh.Signer{hostSigner},
		sshutils.AuthMethods{
			Password:  tunnel.passwordAuth,
			PublicKey: tunnel.keyAuth,
		},
		sshutils.SetLimiter(tunnel.limiter),
	)
	if err != nil {
		return nil, err
	}
	tunnel.userCertChecker = ssh.CertChecker{IsAuthority: tunnel.isUserAuthority}
	tunnel.hostCertChecker = ssh.CertChecker{IsAuthority: tunnel.isHostAuthority}
	return tunnel, nil
}

  /// --- lib/sshutils/server.go --- ///
  func NewServer(
  	component string,
  	a utils.NetAddr,
  	h NewChanHandler,
  	hostSigners []ssh.Signer,
  	ah AuthMethods,
  	opts ...ServerOption) (*Server, error) {

  	err := checkArguments(a, h, hostSigners, ah)
  	if err != nil {
  		return nil, err
  	}
  	s := &Server{
  		component:      component,
  		addr:           a,
  		newChanHandler: h,
  		closeC:         make(chan struct{}),
  	}
  	s.limiter, err = limiter.NewLimiter(limiter.LimiterConfig{})
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}

  	for _, o := range opts {
  		if err := o(s); err != nil {
  			return nil, err
  		}
  	}
  	for _, signer := range hostSigners {
  		(&s.cfg).AddHostKey(signer)
  	}
  	s.cfg.PublicKeyCallback = ah.PublicKey
  	s.cfg.PasswordCallback = ah.Password
  	s.cfg.NoClientAuth = ah.NoClient
  	return s, nil
  }

    /// --- lib/auth/tun.go --- ///
    func (s *AuthTunnel) keyAuth(
    	conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

    	log.Infof("[AUTH] keyAuth: %v->%v, user=%v", conn.RemoteAddr(), conn.LocalAddr(), conn.User())
    	cert, ok := key.(*ssh.Certificate)
    	if !ok {
    		return nil, trace.Errorf("ERROR: Server doesn't support provided key type")
    	}

    	if cert.CertType == ssh.HostCert {
    		err := s.hostCertChecker.CheckHostKey(conn.User(), conn.RemoteAddr(), key)
    		if err != nil {
    			log.Warningf("conn(%v->%v, user=%v) ERROR: failed auth user %v, err: %v",
    				conn.RemoteAddr(), conn.LocalAddr(), conn.User(), conn.User(), err)
    			return nil, err
    		}
    		if err := s.hostCertChecker.CheckCert(conn.User(), cert); err != nil {
    			log.Warningf("conn(%v->%v, user=%v) ERROR: Failed to authorize user %v, err: %v",
    				conn.RemoteAddr(), conn.LocalAddr(), conn.User(), conn.User(), err)
    			return nil, trace.Wrap(err)
    		}
    		perms := &ssh.Permissions{
    			Extensions: map[string]string{
    				ExtHost: conn.User(),
    				ExtRole: cert.Permissions.Extensions[utils.CertExtensionRole],
    			},
    		}
    		return perms, nil
    	}
    	// we are assuming that this is a user cert
    	if err := s.userCertChecker.CheckCert(conn.User(), cert); err != nil {
    		log.Warningf("conn(%v->%v, user=%v) ERROR: Failed to authorize user %v, err: %v",
    			conn.RemoteAddr(), conn.LocalAddr(), conn.User(), conn.User(), err)
    		return nil, trace.Wrap(err)
    	}
    	// we are not using cert extensions for User certificates because of OpenSSH bug
    	// https://bugzilla.mindrot.org/show_bug.cgi?id=2387
    	perms := &ssh.Permissions{
    		Extensions: map[string]string{
    			ExtHost: conn.User(),
    			ExtRole: string(teleport.RoleUser),
    		},
    	}
    	return perms, nil
    }
