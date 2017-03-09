func (n *nauth) GenerateUserCert(pkey, key []byte, teleportUsername string, allowedLogins []string, ttl time.Duration) ([]byte, error) {
	if (ttl > defaults.MaxCertDuration) || (ttl < defaults.MinCertDuration) {
		return nil, trace.BadParameter("wrong certificate TTL")
	}
	if len(allowedLogins) == 0 {
		return nil, trace.BadParameter("allowedLogins: need allowed OS logins")
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return nil, err
	}
	validBefore := uint64(ssh.CertTimeInfinity)
	if ttl != 0 {
		b := time.Now().Add(ttl)
		validBefore = uint64(b.Unix())
		log.Infof("generated user key for %v with expiry on (%v) %v", allowedLogins, validBefore, b)
	}
	// we do not use any extensions in users certs because of this:
	// https://bugzilla.mindrot.org/show_bug.cgi?id=2387
	cert := &ssh.Certificate{
		KeyId:           teleportUsername, // we have to use key id to identify teleport user
		ValidPrincipals: allowedLogins,
		Key:             pubKey,
		ValidBefore:     validBefore,
		CertType:        ssh.UserCert,
	}
	cert.Permissions.Extensions = map[string]string{
		"permit-pty":             "",
		"permit-port-forwarding": "",
	}

	signer, err := ssh.ParsePrivateKey(pkey)
	if err != nil {
		return nil, err
	}
	if err := cert.SignCert(rand.Reader, signer); err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(cert), nil
}



// NewTeleport takes the daemon configuration, instantiates all required services
// and starts them under a supervisor, returning the supervisor object
func NewNodeProcess(cfg *service.PocketConfig) (*PocketNodeProcess, error) {
    var err error
    process := &PocketNodeProcess{
        Supervisor: service.NewSupervisor(),
        Config:     cfg,
    }

    err = process.initSSH();
    if err != nil {
        return nil, err
    }

    return process, nil
}

	func (process *TeleportProcess) initSSH() error {
		process.RegisterWithAuthServer(process.Config.Token, teleport.RoleNode, SSHIdentityEvent)
		eventsC := make(chan Event)
		process.WaitForEvent(SSHIdentityEvent, eventsC, make(chan struct{}))

		var s *srv.Server

		process.RegisterFunc(func() error {
			event := <-eventsC
			log.Infof("[SSH] received %v", &event)
			conn, ok := (event.Payload).(*Connector)
			if !ok {
				return trace.BadParameter("unsupported connector type: %T", event.Payload)
			}

			cfg := process.Config

			limiter, err := limiter.NewLimiter(cfg.SSH.Limiter)
			if err != nil {
				return trace.Wrap(err)
			}

			s, err = srv.New(cfg.SSH.Addr,
				cfg.Hostname,
				[]ssh.Signer{conn.Identity.KeySigner},
				conn.Client,
				cfg.DataDir,
				cfg.AdvertiseIP,
				srv.SetLimiter(limiter),
				srv.SetShell(cfg.SSH.Shell),
				srv.SetAuditLog(conn.Client),
				srv.SetSessionServer(conn.Client),
				srv.SetLabels(cfg.SSH.Labels, cfg.SSH.CmdLabels),
			)
			if err != nil {
				return trace.Wrap(err)
			}

			utils.Consolef(cfg.Console, "[SSH]   Service is starting on %v", cfg.SSH.Addr.Addr)
			if err := s.Start(); err != nil {
				utils.Consolef(cfg.Console, "[SSH]   Error: %v", err)
				return trace.Wrap(err)
			}
			s.Wait()
			log.Infof("[SSH] node service exited")
			return nil
		})
		// execute this when process is asked to exit:
		process.onExit(func(payload interface{}) {
			s.Close()
		})
		return nil
	}

		// RegisterWithAuthServer uses one time provisioning token obtained earlier
		// from the server to get a pair of SSH keys signed by Auth server host
		// certificate authority
		func (process *TeleportProcess) RegisterWithAuthServer(token string, role teleport.Role, eventName string) {
			cfg := process.Config
			identityID := auth.IdentityID{Role: role, HostUUID: cfg.HostUUID}

			// this means the server has not been initialized yet, we are starting
			// the registering client that attempts to connect to the auth server
			// and provision the keys
			var authClient *auth.TunClient
			process.RegisterFunc(func() error {
				retryTime := defaults.ServerHeartbeatTTL / 3
				for {
					connector, err := process.connectToAuthService(role)
					if err == nil {
						process.BroadcastEvent(Event{Name: eventName, Payload: connector})
						authClient = connector.Client
						return nil
					}
					if trace.IsConnectionProblem(err) {
						utils.Consolef(cfg.Console, "[%v] connecting to auth server: %v", role, err)
						time.Sleep(retryTime)
						continue
					}
					if !trace.IsNotFound(err) {
						return trace.Wrap(err)
					}
					//  we haven't connected yet, so we expect the token to exist
					if process.getLocalAuth() != nil {
						// Auth service is on the same host, no need to go though the invitation
						// procedure
						log.Debugf("[Node] this server has local Auth server started, using it to add role to the cluster")
						err = auth.LocalRegister(cfg.DataDir, identityID, process.getLocalAuth())
					} else {
						// Auth server is remote, so we need a provisioning token
						if token == "" {
							return trace.BadParameter("%v must join a cluster and needs a provisioning token", role)
						}
						log.Infof("[Node] %v joining the cluster with a token %v", role, token)
						err = auth.Register(cfg.DataDir, token, identityID, cfg.AuthServers)
					}
					if err != nil {
						utils.Consolef(cfg.Console, "[%v] failed to join the cluster: %v", role, err)
						time.Sleep(retryTime)
					} else {
						utils.Consolef(cfg.Console, "[%v] Successfully registered with the cluster", role)
						continue
					}
				}
			})

			process.onExit(func(interface{}) {
				if authClient != nil {
					authClient.Close()
				}
			})
		}

			// *** THIS IS WHERE CERTIFICATE AND KEY IS REGISTERED ***

			// Register is used by auth service clients (other services, like proxy or SSH) when a new node
			// joins the cluster
			func Register(dataDir, token string, id IdentityID, servers []utils.NetAddr) error {
				tok, err := readToken(token)
				if err != nil {
					return trace.Wrap(err)
				}
				method, err := NewTokenAuth(id.HostUUID, tok)
				if err != nil {
					return trace.Wrap(err)
				}

				client, err := NewTunClient(
					"auth.client.register",
					servers,
					id.HostUUID,
					method)
				if err != nil {
					return trace.Wrap(err)
				}
				defer client.Close()

				keys, err := client.RegisterUsingToken(tok, id.HostUUID, id.Role)
				if err != nil {
					return trace.Wrap(err)
				}
				return writeKeys(dataDir, id, keys.Key, keys.Cert)
			}

				// RegisterUserToken calls the auth service API to register a new node via registration token
				// which has been previously issued via GenerateToken
				func (c *Client) RegisterUsingToken(token, hostID string, role teleport.Role) (*PackedKeys, error) {
					out, err := c.PostJSON(c.Endpoint("tokens", "register"),
						registerUsingTokenReq{
							HostID: hostID,
							Token:  token,
							Role:   role,
						})
					if err != nil {
						return nil, trace.Wrap(err)
					}
					var keys PackedKeys
					if err := json.Unmarshal(out.Bytes(), &keys); err != nil {
						return nil, trace.Wrap(err)
					}
					return &keys, nil
				}


			// "/tokens/register"
			type PackedKeys struct {
				Key  []byte `json:"key"`
				Cert []byte `json:"cert"`
			}
			keys, err := client.RegisterUsingToken(tok, id.HostUUID, id.Role)

/// server side
srv.POST("/v1/tokens/register", httplib.MakeHandler(srv.registerUsingToken))

type registerUsingTokenReq struct {
	HostID string        `json:"hostID"`
	Role   teleport.Role `json:"role"`
	Token  string        `json:"token"`
}

func (s *APIServer) registerUsingToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	var req *registerUsingTokenReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}
	keys, err := s.a.RegisterUsingToken(req.Token, req.HostID, req.Role)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return keys, nil
}

	func (a *AuthWithRoles) RegisterUsingToken(token, hostID string, role teleport.Role) (*PackedKeys, error) {
		if err := a.permChecker.HasPermission(a.role, ActionRegisterUsingToken); err != nil {
			return nil, trace.Wrap(err)
		}
		return a.authServer.RegisterUsingToken(token, hostID, role)

	}

		// RegisterUsingToken adds a new node to the Teleport cluster using previously issued token.
		// A node must also request a specific role (and the role must match one of the roles
		// the token was generated for).
		//
		// If a token was generated with a TTL, it gets enforced (can't register new nodes after TTL expires)
		// If a token was generated with a TTL=0, it means it's a single-use token and it gets destroyed
		// after a successful registration.
		func (s *AuthServer) RegisterUsingToken(token, hostID string, role teleport.Role) (*PackedKeys, error) {
			log.Infof("[AUTH] Node `%v` is trying to join as %v", hostID, role)
			if hostID == "" {
				return nil, trace.BadParameter("HostID cannot be empty")
			}
			if err := role.Check(); err != nil {
				return nil, trace.Wrap(err)
			}
			// make sure the token is valid:
			roles, err := s.ValidateToken(token)
			if err != nil {
				msg := fmt.Sprintf("`%v` cannot join the cluster as %s. Token error: %v", hostID, role, err)
				log.Warnf("[AUTH] %s", msg)
				return nil, trace.AccessDenied(msg)
			}
			// make sure the caller is requested wthe role allowed by the token:
			if !roles.Include(role) {
				msg := fmt.Sprintf("'%v' cannot join the cluster, the token does not allow '%s' role", hostID, role)
				log.Warningf("[AUTH] %s", msg)
				return nil, trace.BadParameter(msg)
			}
			if !s.checkTokenTTL(token) {
				return nil, trace.AccessDenied("'%v' cannot join the cluster. The token has expired", hostID)
			}
			// generate & return the node cert:
			keys, err := s.GenerateServerKeys(hostID, teleport.Roles{role})
			if err != nil {
				return nil, trace.Wrap(err)
			}
			utils.Consolef(os.Stdout, "[AUTH] Node `%v` joined the cluster", hostID)
			return keys, nil
		}

			// GenerateServerKeys generates private key and certificate signed
			// by the host certificate authority, listing the role of this server
			func (s *AuthServer) GenerateServerKeys(hostID string, roles teleport.Roles) (*PackedKeys, error) {
				k, pub, err := s.GenerateKeyPair("")
				if err != nil {
					return nil, trace.Wrap(err)
				}
				// we always append authority's domain to resulting node name,
				// that's how we make sure that nodes are uniquely identified/found
				// in cases when we have multiple environments/organizations
				fqdn := fmt.Sprintf("%s.%s", hostID, s.DomainName)
				c, err := s.GenerateHostCert(pub, fqdn, s.DomainName, roles, 0)
				if err != nil {
					log.Warningf("[AUTH] Node `%v` cannot join: cert generation error. %v", hostID, err)
					return nil, trace.Wrap(err)
				}

				return &PackedKeys{
					Key:  k,
					Cert: c,
				}, nil
			}

				// GenerateKeyPair returns fresh priv/pub keypair, takes about 300ms to execute
				func (n *nauth) GenerateKeyPair(passphrase string) ([]byte, []byte, error) {
					priv, err := rsa.GenerateKey(rand.Reader, 2048)
					if err != nil {
						return nil, nil, err
					}
					privDer := x509.MarshalPKCS1PrivateKey(priv)
					privBlock := pem.Block{
						Type:    "RSA PRIVATE KEY",
						Headers: nil,
						Bytes:   privDer,
					}
					privPem := pem.EncodeToMemory(&privBlock)

					pub, err := ssh.NewPublicKey(&priv.PublicKey)
					if err != nil {
						return nil, nil, err
					}
					pubBytes := ssh.MarshalAuthorizedKey(pub)
					return privPem, pubBytes, nil
				}

				//// c, err := s.GenerateHostCert(pub, fqdn, s.DomainName, roles, 0)

				// GenerateHostCert generates host certificate, it takes pkey as a signing
				// private key (host certificate authority)
				func (s *AuthServer) GenerateHostCert(key []byte, hostID, authDomain string, roles teleport.Roles, ttl time.Duration) ([]byte, error) {
					ca, err := s.Trust.GetCertAuthority(services.CertAuthID{
						Type:       services.HostCA,
						DomainName: s.DomainName,
					}, true)
					if err != nil {
						return nil, trace.Errorf("failed to load host CA for '%s': %v", s.DomainName, err)
					}
					privateKey, err := ca.FirstSigningKey()
					if err != nil {
						return nil, trace.Wrap(err)
					}
					return s.Authority.GenerateHostCert(privateKey, key, hostID, authDomain, roles, ttl)
				}

					// *** lib/services/local/trust.go ***
					// CertAuthID - id of certificate authority (it's type and domain name)
					type CertAuthID struct {
						Type       CertAuthType `json:"type"`
						DomainName string       `json:"domain_name"`
					}

					// CertAuthority is a host or user certificate authority that
					// can check and if it has private key stored as well, sign it too
					type CertAuthority struct {
						// Type is either user or host certificate authority
						Type CertAuthType `json:"type"`
						// DomainName identifies domain name this authority serves,
						// for host authorities that means base hostname of all servers,
						// for user authorities that means organization name
						DomainName string `json:"domain_name"`
						// Checkers is a list of SSH public keys that can be used to check
						// certificate signatures
						CheckingKeys [][]byte `json:"checking_keys"`
						// SigningKeys is a list of private keys used for signing
						SigningKeys [][]byte `json:"signing_keys"`
						// AllowedLogins is a list of allowed logins for users within
						// this certificate authority
						AllowedLogins []string `json:"allowed_logins"`
					}

					// GetCertAuthority returns certificate authority by given id. Parameter loadSigningKeys
					// controls if signing keys are loaded
					func (s *CA) GetCertAuthority(id services.CertAuthID, loadSigningKeys bool) (*services.CertAuthority, error) {
						if err := id.Check(); err != nil {
							return nil, trace.Wrap(err)
						}
						val, err := s.backend.GetVal([]string{"authorities", string(id.Type)}, id.DomainName)
						if err != nil {
							return nil, trace.Wrap(err)
						}
						var ca *services.CertAuthority
						if err := json.Unmarshal(val, &ca); err != nil {
							return nil, trace.Wrap(err)
						}
						if err := ca.Check(); err != nil {
							return nil, trace.Wrap(err)
						}
						if !loadSigningKeys {
							ca.SigningKeys = nil
						}
						return ca, nil
					}

					func (n *nauth) GenerateHostCert(privateSigningKey, publicKey []byte, hostname, authDomain string, roles teleport.Roles, ttl time.Duration) ([]byte, error) {
						if err := roles.Check(); err != nil {
							return nil, trace.Wrap(err)
						}
						pubKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
						if err != nil {
							return nil, err
						}
						validBefore := uint64(ssh.CertTimeInfinity)
						if ttl != 0 {
							b := time.Now().Add(ttl)
							validBefore = uint64(b.Unix())
						}
						cert := &ssh.Certificate{
							ValidPrincipals: []string{hostname},
							Key:             pubKey,
							ValidBefore:     validBefore,
							CertType:        ssh.HostCert,
						}
						cert.Permissions.Extensions = make(map[string]string)
						cert.Permissions.Extensions[utils.CertExtensionRole] = roles.String()
						cert.Permissions.Extensions[utils.CertExtensionAuthority] = string(authDomain)
						signer, err := ssh.ParsePrivateKey(privateSigningKey)
						if err != nil {
							return nil, err
						}
						if err := cert.SignCert(rand.Reader, signer); err != nil {
							return nil, err
						}
						return ssh.MarshalAuthorizedKey(cert), nil
					}
