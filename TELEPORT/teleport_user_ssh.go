/// --- tool/tsh/main.go --- ///
// onSSH executes 'tsh ssh' command
func onSSH(cf *CLIConf) {
	tc, err := makeClient(cf, false)
	if err != nil {
		utils.FatalError(err)
	}

	tc.Stdin = os.Stdin
	if err = tc.SSH(context.TODO(), cf.RemoteCommand, cf.LocalExec); err != nil {
		// exit with the same exit status as the failed command:
		if tc.ExitStatus != 0 {
			os.Exit(tc.ExitStatus)
		} else {
			utils.FatalError(err)
		}
	}
}

	// makeClient takes the command-line configuration and constructs & returns
	// a fully configured TeleportClient object
	func makeClient(cf *CLIConf) (tc *client.TeleportClient, err error) {
		// apply defults
		if cf.NodePort == 0 {
			cf.NodePort = defaults.SSHServerListenPort
		}
		if cf.MinsToLive == 0 {
			cf.MinsToLive = int32(defaults.CertDuration / time.Minute)
		}

		// split login & host
		hostLogin := cf.NodeLogin
		var labels map[string]string
		if cf.UserHost != "" {
			parts := strings.Split(cf.UserHost, "@")
			if len(parts) > 1 {
				hostLogin = parts[0]
				cf.UserHost = parts[1]
			}
			// see if remote host is specified as a set of labels
			if strings.Contains(cf.UserHost, "=") {
				labels, err = client.ParseLabelSpec(cf.UserHost)
				if err != nil {
					return nil, err
				}
			}
		}
		fPorts, err := client.ParsePortForwardSpec(cf.LocalForwardPorts)
		if err != nil {
			return nil, err
		}

		// prep client config:
		c := &client.Config{
			Stdout:             os.Stdout,
			Stderr:             os.Stderr,
			Stdin:              os.Stdin,
			Username:           cf.Username,
			ProxyHost:          cf.Proxy,
			Host:               cf.UserHost,
			HostPort:           int(cf.NodePort),
			HostLogin:          hostLogin,
			Labels:             labels,
			KeyTTL:             time.Minute * time.Duration(cf.MinsToLive),
			InsecureSkipVerify: cf.InsecureSkipVerify,
			LocalForwardPorts:  fPorts,
			ConnectorID:        cf.ExternalAuth,
			SiteName:           cf.SiteName,
			Interactive:        cf.Interactive,
		}
		return client.NewClient(c)
	}

# -- BEGIN TeleportClient.SSH(context.Context, []string, bool) error --

	/// --- tool/tsh/main.go --- ///
	// SSH connects to a node and, if 'command' is specified, executes the command on it,
	// otherwise runs interactive shell
	//
	// Returns nil if successful, or (possibly) *exec.ExitError
	func (tc *TeleportClient) SSH(ctx context.Context, command []string, runLocally bool) error {
		// connect to proxy first:
		if !tc.Config.ProxySpecified() {
			return trace.BadParameter("proxy server is not specified")
		}
		proxyClient, err := tc.ConnectToProxy()
		if err != nil {
			return trace.Wrap(err)
		}
		defer proxyClient.Close()
		siteInfo, err := proxyClient.currentSite()
		if err != nil {
			return trace.Wrap(err)
		}
		// which nodes are we executing this commands on?
		nodeAddrs, err := tc.getTargetNodes(ctx, proxyClient)
		if err != nil {
			return trace.Wrap(err)
		}
		if len(nodeAddrs) == 0 {
			return trace.BadParameter("no target host specified")
		}
		// more than one node for an interactive shell?
		// that can't be!
		if len(nodeAddrs) != 1 {
			fmt.Printf(
				"\x1b[1mWARNING\x1b[0m: multiple nodes match the label selector. Picking %v (first)\n",
				nodeAddrs[0])
		}
		nodeClient, err := proxyClient.ConnectToNode(
			ctx,
			nodeAddrs[0]+"@"+siteInfo.Name,
			tc.Config.HostLogin,
			false)
		if err != nil {
			tc.ExitStatus = 1
			return trace.Wrap(err)
		}
		// proxy local ports (forward incoming connections to remote host ports)
		tc.startPortForwarding(nodeClient)

		// local execution?
		if runLocally {
			if len(tc.Config.LocalForwardPorts) == 0 {
				fmt.Println("Executing command locally without connecting to any servers. This makes no sense.")
			}
			return runLocalCommand(command)
		}
		// execute command(s) or a shell on remote node(s)
		if len(command) > 0 {
			return tc.runCommand(ctx, siteInfo.Name, nodeAddrs, proxyClient, command)
		}
		return tc.runShell(nodeClient, nil)
	}

		/// --- lib/client/api.go --- ///
		// ConnectToProxy dials the proxy server and returns ProxyClient if successful
		func (tc *TeleportClient) ConnectToProxy() (*ProxyClient, error) {
			proxyAddr := tc.Config.ProxySSHHostPort()
			sshConfig := &ssh.ClientConfig{
				User:            tc.getProxySSHPrincipal(),
				HostKeyCallback: tc.HostKeyCallback,
			}
			// helper to create a ProxyClient struct
			makeProxyClient := func(sshClient *ssh.Client, m ssh.AuthMethod) *ProxyClient {
				return &ProxyClient{
					Client:          sshClient,
					proxyAddress:    proxyAddr,
					hostKeyCallback: sshConfig.HostKeyCallback,
					authMethod:      m,
					hostLogin:       tc.Config.HostLogin,
					siteName:        tc.Config.SiteName,
				}
			}
			successMsg := fmt.Sprintf("[CLIENT] successful auth with proxy %v", proxyAddr)
			// try to authenticate using every non interactive auth method we have:
			for i, m := range tc.authMethods() {
				log.Infof("[CLIENT] connecting proxy=%v login='%v' method=%d", proxyAddr, sshConfig.User, i)

				sshConfig.Auth = []ssh.AuthMethod{m}
				sshClient, err := ssh.Dial("tcp", proxyAddr, sshConfig)
				if err != nil {
					if utils.IsHandshakeFailedError(err) {
						log.Warn(err)
						continue
					}
					return nil, trace.Wrap(err)
				}
				log.Infof(successMsg)
				return makeProxyClient(sshClient, m), nil
			}
			// we have exhausted all auth existing auth methods and local login
			// is disabled in configuration
			if tc.Config.SkipLocalAuth {
				return nil, trace.BadParameter("failed to authenticate with proxy %v", proxyAddr)
			}
			// if we get here, it means we failed to authenticate using stored keys
			// and we need to ask for the login information
			authMethod, err := tc.Login()
			if err != nil {
				// we need to communicate directly to user here,
				// otherwise user will see endless loop with no explanation
				if trace.IsTrustError(err) {
					fmt.Printf("Refusing to connect to untrusted proxy %v without --insecure flag\n", proxyAddr)
				}
				return nil, trace.Wrap(err)
			}

			// After successfull login we have local agent updated with latest
			// and greatest auth information, try it now
			sshConfig.Auth = []ssh.AuthMethod{authMethod}
			sshConfig.User = tc.getProxySSHPrincipal()
			sshClient, err := ssh.Dial("tcp", proxyAddr, sshConfig)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			log.Debugf(successMsg)
			proxyClient := makeProxyClient(sshClient, authMethod)
			// get (and remember) the site info:
			site, err := proxyClient.currentSite()
			if err != nil {
				return nil, trace.Wrap(err)
			}
			tc.SiteName = site.Name
			return proxyClient, nil
		}

			/// --- lib/client/api.go --- ///
			// authMethods returns a list (slice) of all SSH auth methods this client
			// can use to try to authenticate
			func (tc *TeleportClient) authMethods() []ssh.AuthMethod {
				// return the auth methods that we were configured with
				// plus our local key agent (i.e. methods we've added during runtime
				// by the means of .AddKey())
				m := append([]ssh.AuthMethod(nil), tc.Config.AuthMethods...)
				return append(m, tc.LocalAgent().AuthMethods()...)
			}

			/// --- lib/client/keyagent.go --- ///
			// AuthMethods returns the list of differnt authentication methods this agent supports
			// It returns two:
			//	  1. First to try is the external SSH agent
			//    2. Itself (disk-based local agent)
			func (a *LocalKeyAgent) AuthMethods() (m []ssh.AuthMethod) {
				// combine our certificates with external SSH agent's:
				var certs []ssh.Signer
				if ourCerts, _ := a.Signers(); ourCerts != nil {
					certs = append(certs, ourCerts...)
				}
				if a.sshAgent != nil {
					if sshAgentCerts, _ := a.sshAgent.Signers(); sshAgentCerts != nil {
						certs = append(certs, sshAgentCerts...)
					}
				}
				// for every certificate create a new "auth method" and return them
				m = make([]ssh.AuthMethod, len(certs))
				for i := range certs {
					m[i] = methodForCert(certs[i])
				}
				return m
			}

				/// --- lib/client/keyagent.go --- ///
				// CertAuthMethod is a wrapper around ssh.Signer (certificate signer) object.
				// CertAuthMethod then implements ssh.Authmethod interface around this one certificate signer.
				//
				// We need this wrapper because Golang's SSH library's unfortunate API design. It uses
				// callbacks with 'authMethod' interfaces and without this wrapper it is impossible to
				// tell which certificate an 'authMethod' passed via a callback had succeeded authenticating with.
				type CertAuthMethod struct {
					ssh.AuthMethod
					Cert ssh.Signer
				}

				func methodForCert(cert ssh.Signer) *CertAuthMethod {
					return &CertAuthMethod{
						Cert: cert,
						AuthMethod: ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
							return []ssh.Signer{cert}, nil
						}),
					}
				}

		/// --- lib/client/client.go --- ///
		// ConnectToNode connects to the ssh server via Proxy.
		// It returns connected and authenticated NodeClient
		func (proxy *ProxyClient) ConnectToNode(ctx context.Context, nodeAddress string, user string, quiet bool) (*NodeClient, error) {
			log.Infof("[CLIENT] connecting to node: %s", nodeAddress)

			// parse destination first:
			localAddr, err := utils.ParseAddr("tcp://" + proxy.proxyAddress)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			fakeAddr, err := utils.ParseAddr("tcp://" + nodeAddress)
			if err != nil {
				return nil, trace.Wrap(err)
			}

			// we have to try every auth method separatedly,
			// because go SSH will try only one (fist) auth method
			// of a given type, so if you have 2 different public keys
			// you have to try each one differently
			proxySession, err := proxy.Client.NewSession()
			if err != nil {
				return nil, trace.Wrap(err)
			}
			proxyWriter, err := proxySession.StdinPipe()
			if err != nil {
				return nil, trace.Wrap(err)
			}
			proxyReader, err := proxySession.StdoutPipe()
			if err != nil {
				return nil, trace.Wrap(err)
			}
			proxyErr, err := proxySession.StderrPipe()
			if err != nil {
				return nil, trace.Wrap(err)
			}
			err = proxySession.RequestSubsystem("proxy:" + nodeAddress)
			if err != nil {
				// read the stderr output from the failed SSH session and append
				// it to the end of our own message:
				serverErrorMsg, _ := ioutil.ReadAll(proxyErr)
				return nil, trace.Errorf("failed connecting to node %v. %s",
					nodeName(strings.Split(nodeAddress, "@")[0]), serverErrorMsg)
			}
			pipeNetConn := utils.NewPipeNetConn(
				proxyReader,
				proxyWriter,
				proxySession,
				localAddr,
				fakeAddr,
			)
			sshConfig := &ssh.ClientConfig{
				User:            user,
				Auth:            []ssh.AuthMethod{proxy.authMethod},
				HostKeyCallback: proxy.hostKeyCallback,
			}
			conn, chans, reqs, err := newClientConn(ctx, pipeNetConn, nodeAddress, sshConfig)
			if err != nil {
				if utils.IsHandshakeFailedError(err) {
					proxySession.Close()
					parts := strings.Split(nodeAddress, "@")
					hostname := parts[0]
					if len(hostname) == 0 && len(parts) > 1 {
						hostname = "cluster " + parts[1]
					}
					return nil, trace.Errorf(`access denied to %v connecting to %v`, user, nodeName(hostname))
				}
				return nil, trace.Wrap(err)
			}

			client := ssh.NewClient(conn, chans, reqs)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			return &NodeClient{Client: client, Proxy: proxy}, nil
		}

		/// --- lib/client/api.go --- ///
		func (tc *TeleportClient) startPortForwarding(nodeClient *NodeClient) error {
			if len(tc.Config.LocalForwardPorts) > 0 {
				for _, fp := range tc.Config.LocalForwardPorts {
					socket, err := net.Listen("tcp", net.JoinHostPort(fp.SrcIP, strconv.Itoa(fp.SrcPort)))
					if err != nil {
						return trace.Wrap(err)
					}
					go nodeClient.listenAndForward(socket, net.JoinHostPort(fp.DestHost, strconv.Itoa(fp.DestPort)))
				}
			}
			return nil
		}

# -- END TeleportClient.SSH(context.Context, []string, bool) error --

			/// --- lib/client/api.go --- ///
			// Login logs the user into a Teleport cluster by talking to a Teleport proxy.
			// If successful, saves the received session keys into the local keystore for future use.
			func (tc *TeleportClient) Login() (*CertAuthMethod, error) {
				// generate a new keypair. the public key will be signed via proxy if our password+HOTP  are legit
				key, err := tc.MakeKey()
				if err != nil {
					return nil, trace.Wrap(err)
				}
				var response *web.SSHLoginResponse
				if tc.ConnectorID == "" {
					response, err = tc.directLogin(key.Pub)
					if err != nil {
						return nil, trace.Wrap(err)
					}
				} else {
					response, err = tc.oidcLogin(tc.ConnectorID, key.Pub)
					if err != nil {
						return nil, trace.Wrap(err)
					}
					// in this case identity is returned by the proxy
					tc.Username = response.Username
				}
				key.Cert = response.Cert

				// save the list of CAs we trust to the cache file
				err = tc.localAgent.AddHostSignersToCache(response.HostSigners)
				if err != nil {
					return nil, trace.Wrap(err)
				}

				// save the key:
				return tc.localAgent.AddKey(tc.ProxyHost(), tc.Config.Username, key)
			}

			/// --- lib/client/api.go --- ///
			// directLogin asks for a password + HOTP token, makes a request to CA via proxy
			func (tc *TeleportClient) directLogin(pub []byte) (*web.SSHLoginResponse, error) {
				httpsProxyHostPort := tc.Config.ProxyWebHostPort()
				certPool := loopbackPool(httpsProxyHostPort)

				// ping the HTTPs endpoint first:
				if err := web.Ping(httpsProxyHostPort, tc.InsecureSkipVerify, certPool); err != nil {
					return nil, trace.Wrap(err)
				}

				password, hotpToken, err := tc.AskPasswordAndHOTP()
				if err != nil {
					return nil, trace.Wrap(err)
				}

				// ask the CA (via proxy) to sign our public key:
				response, err := web.SSHAgentLogin(httpsProxyHostPort,
					tc.Config.Username,
					password,
					hotpToken,
					pub,
					tc.KeyTTL,
					tc.InsecureSkipVerify,
					certPool)

				return response, trace.Wrap(err)
			}

					/// --- lib/client/api.go --- ///
					// AskPasswordAndHOTP prompts the user to enter the password + HTOP 2nd factor
					func (tc *TeleportClient) AskPasswordAndHOTP() (pwd string, token string, err error) {
						fmt.Printf("Enter password for Teleport user %v:\n", tc.Config.Username)
						pwd, err = passwordFromConsole()
						if err != nil {
							fmt.Println(err)
							return "", "", trace.Wrap(err)
						}

						fmt.Printf("Enter your HOTP token:\n")
						token, err = lineFromConsole()
						if err != nil {
							fmt.Println(err)
							return "", "", trace.Wrap(err)
						}
						return pwd, token, nil
					}

					/// --- lib/client/sshlogin.go --- ///
					// SSHAgentLogin issues call to web proxy and receives temp certificate
					// if credentials are valid
					//
					// proxyAddr must be specified as host:port
					func SSHAgentLogin(proxyAddr, user, password, hotpToken string, pubKey []byte, ttl time.Duration, insecure bool, pool *x509.CertPool) (*SSHLoginResponse, error) {
						clt, _, err := initClient(proxyAddr, insecure, pool)
						if err != nil {
							return nil, trace.Wrap(err)
						}
						re, err := clt.PostJSON(clt.Endpoint("webapi", "ssh", "certs"), createSSHCertReq{
							User:      user,
							Password:  password,
							HOTPToken: hotpToken,
							PubKey:    pubKey,
							TTL:       ttl,
						})
						if err != nil {
							return nil, trace.Wrap(err)
						}

						var out *SSHLoginResponse
						err = json.Unmarshal(re.Bytes(), &out)
						if err != nil {
							return nil, trace.Wrap(err)
						}

						return out, nil
					}

// -- SERVER -- //
/// --- lib/web/web.go --- ///
// Issues SSH temp certificates based on 2FA access creds
h.POST("/webapi/ssh/certs", httplib.MakeHandler(h.createSSHCert))

// createSSHCertReq are passed by web client
// to authenticate against teleport server and receive
// a temporary cert signed by auth server authority
type createSSHCertReq struct {
	// User is a teleport username
	User string `json:"user"`
	// Password is user's pass
	Password string `json:"password"`
	// HOTPToken is second factor token
	HOTPToken string `json:"hotp_token"`
	// PubKey is a public key user wishes to sign
	PubKey []byte `json:"pub_key"`
	// TTL is a desired TTL for the cert (max is still capped by server,
	// however user can shorten the time)
	TTL time.Duration `json:"ttl"`
}

/// --- lib/web/web.go --- ///
// SSHLoginResponse is a response returned by web proxy
type SSHLoginResponse struct {
	// User contains a logged in user informationn
	Username string `json:"username"`
	// Cert is a signed certificate
	Cert []byte `json:"cert"`
	// HostSigners is a list of signing host public keys
	// trusted by proxy
	HostSigners []services.CertAuthority `json:"host_signers"`
}

// createSSHCert is a web call that generates new SSH certificate based
// on user's name, password, 2nd factor token and public key user wishes to sign
//
// POST /v1/webapi/ssh/certs
//
// { "user": "bob", "password": "pass", "hotp_token": "tok", "pub_key": "key to sign", "ttl": 1000000000 }
//
// Success response
//
// { "cert": "base64 encoded signed cert", "host_signers": [{"domain_name": "example.com", "checking_keys": ["base64 encoded public signing key"]}] }
//
func (h *Handler) createSSHCert(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	var req *createSSHCertReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}

	cert, err := h.auth.GetCertificate(*req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return cert, nil
}

	/// --- lib/web/web.go --- ///
	// SSHLoginResponse is a response returned by web proxy
	type SSHLoginResponse struct {
		// User contains a logged in user informationn
		Username string `json:"username"`
		// Cert is a signed certificate
		Cert []byte `json:"cert"`
		// HostSigners is a list of signing host public keys
		// trusted by proxy
		HostSigners []services.CertAuthority `json:"host_signers"`
	}

	/// --- lib/web/sessions.go --- ///
	func (s *sessionCache) GetCertificate(c createSSHCertReq) (*SSHLoginResponse, error) {
		method, err := auth.NewWebPasswordAuth(c.User, []byte(c.Password),
			c.HOTPToken)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		clt, err := auth.NewTunClient("web.session", s.authServers, c.User, method)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		defer clt.Close()
		cert, err := clt.GenerateUserCert(c.PubKey, c.User, c.TTL)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		hostSigners, err := clt.GetCertAuthorities(services.HostCA, false)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		signers := []services.CertAuthority{}
		for _, hs := range hostSigners {
			signers = append(signers, *hs)
		}

		return &SSHLoginResponse{
			Cert:        cert,
			HostSigners: signers,
		}, nil
	}

		/// --- lib/auth/tun.go --- ///
		// authBucket uses password to transport app-specific user name and
		// auth-type in addition to the password to support auth
		type authBucket struct {
			User      string `json:"user"`
			Type      string `json:"type"`
			Pass      []byte `json:"pass"`
			HotpToken string `json:"hotpToken"`
		}

		/// --- lib/auth/pc_tun.go --- ///
		func NewWebPasswordAuth(user string, password []byte, hotpToken string) ([]ssh.AuthMethod, error) {
			data, err := json.Marshal(authBucket{
				Type:      AuthWebPassword,
				User:      user,
				Pass:      password,
				HotpToken: hotpToken,
			})
			if err != nil {
				return nil, err
			}
			return []ssh.AuthMethod{ssh.Password(string(data))}, nil
		}

			/// --- lib/auth/pc_tun.go --- ///
			func (s *AuthTunnel) passwordAuth(
			conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
					var ab *authBucket
					if err := json.Unmarshal(password, &ab); err != nil {
							return nil, err
					}

					log.Infof("[AUTH] login attempt: user '%v' type '%v'", conn.User(), ab.Type)

					switch ab.Type {
					case AuthWebPassword:
			        if err := s.authServer.CheckPassword(conn.User(), ab.Pass, ab.HotpToken); err != nil {
			            log.Warningf("password auth error: %#v", err)
			            return nil, trace.Wrap(err)
			        }
			        perms := &ssh.Permissions{
			            Extensions: map[string]string{
			                ExtWebPassword: "<password>",
			                ExtRole:        string(teleport.RoleUser),
			            },
			        }
			        log.Infof("[AUTH] password authenticated user: '%v'", conn.User())
			        return perms, nil
					}
			}

		/// --- lib/auth/apiserver.go --- ///
		type generateUserCertReq struct {
			Key  []byte        `json:"key"`
			User string        `json:"user"`
			TTL  time.Duration `json:"ttl"`
		}

		/// --- lib/auth/clt.go --- //
		// GenerateUserCert takes the public key in the Open SSH ``authorized_keys``
		// plain text format, signs it using User Certificate Authority signing key and returns the
		// resulting certificate.
		func (c *Client) GenerateUserCert(key []byte, user string, ttl time.Duration) ([]byte, error) {
			out, err := c.PostJSON(c.Endpoint("ca", "user", "certs"),
				generateUserCertReq{
					Key:  key,
					User: user,
					TTL:  ttl,
				})
			if err != nil {
				return nil, trace.Wrap(err)
			}
			var cert string
			if err := json.Unmarshal(out.Bytes(), &cert); err != nil {
				return nil, trace.Wrap(err)
			}
			return []byte(cert), nil
		}

// -- BEGIN SERVER -- //
srv.POST("/v1/ca/user/certs", httplib.MakeHandler(srv.generateUserCert))

type generateUserCertReq struct {
	Key  []byte        `json:"key"`
	User string        `json:"user"`
	TTL  time.Duration `json:"ttl"`
}

func (s *APIServer) generateUserCert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	var req *generateUserCertReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		log.Errorf("failed parsing JSON request. %v", err)
		return nil, trace.Wrap(err)
	}
	// SSH-to-HTTP gateway (tun server) sets HTTP basic auth to SSH cert principal
	// This allows us to make sure that users can only request new certificates
	// only for themselves, except admin users
	caller, _, ok := r.BasicAuth()
	if !ok {
		return nil, trace.AccessDenied("missing username or password")
	}
	if req.User != caller && s.a.role != teleport.RoleAdmin {
		return nil, trace.AccessDenied("user %s cannot request a certificate for %s",
			caller, req.User)
	}
	cert, err := s.a.GenerateUserCert(req.Key, req.User, req.TTL)
	if err != nil {
		log.Error(err)
		return nil, trace.Wrap(err)
	}
	return string(cert), nil
}
// -- END SERVER -- //


		func (c *Client) GetCertAuthorities(caType services.CertAuthType, loadKeys bool) ([]*services.CertAuthority, error) {
			if err := caType.Check(); err != nil {
				return nil, trace.Wrap(err)
			}
			out, err := c.Get(c.Endpoint("authorities", string(caType)), url.Values{
				"load_keys": []string{fmt.Sprintf("%t", loadKeys)},
			})
			if err != nil {
				return nil, trace.Wrap(err)
			}
			var re []*services.CertAuthority
			if err := json.Unmarshal(out.Bytes(), &re); err != nil {
				return nil, err
			}
			return re, nil
		}

// -- BEGIN SERVER -- //
srv.GET("/v1/authorities/:type", httplib.MakeHandler(srv.getCertAuthorities))

func (s *APIServer) getCertAuthorities(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	loadKeys, _, err := httplib.ParseBool(r.URL.Query(), "load_keys")
	if err != nil {
		return nil, trace.Wrap(err)
	}
	certs, err := s.a.GetCertAuthorities(services.CertAuthType(p[0].Value), loadKeys)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return certs, nil
}
// -- END SERVER -- //


/*
INFO[0437] [AUTH] login attempt: user 'root' type 'password'
WARN[0437] password auth error:
	&trace.TraceErr{
		Err:(*trace.AccessDeniedError)(0xc420555d50),
		Traces:trace.Traces{
			trace.Trace{
				Path:"/Users/almightykim/GOREPO/src/github.com/gravitational/teleport/lib/services/local/users.go",
				Func:"github.com/gravitational/teleport/lib/services/local.(*IdentityService).CheckPassword", Line:332
			}
		},
		Message:"bad one time token",
		DebugMessage:""
	}

-> access denied to 'root': bad username or credentials
*/


/************************************************ AUTH SERVER ************************************************/
// -- BEGIN SERVER -- //
func (process *TeleportProcess) initAuthService(authority auth.Authority) error {

	// NewTunnel creates a new SSH tunnel server which is not started yet.
	// This is how "site API" (aka "auth API") is served: by creating
	// an "tunnel server" which serves HTTP via SSH.
	func NewTunnel(addr utils.NetAddr,
		hostSigner ssh.Signer,
		apiConf *APIConfig,
		opts ...ServerOption) (tunnel *AuthTunnel, err error) {

		tunnel = &AuthTunnel{
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

		func (s *AuthTunnel) passwordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			var ab *authBucket
			if err := json.Unmarshal(password, &ab); err != nil {
				return nil, err
			}

			log.Infof("[AUTH] login attempt: user '%v' type '%v'", conn.User(), ab.Type)

			switch ab.Type {
			case AuthWebPassword:
				if err := s.authServer.CheckPassword(conn.User(), ab.Pass, ab.HotpToken); err != nil {
					log.Warningf("password auth error: %#v", err)
					return nil, trace.Wrap(err)
				}
				perms := &ssh.Permissions{
					Extensions: map[string]string{
						ExtWebPassword: "<password>",
						ExtRole:        string(teleport.RoleUser),
					},
				}
				log.Infof("[AUTH] password authenticated user: '%v'", conn.User())
				return perms, nil
			}
		}
// -- END SERVER -- //





/************************************************ THIS IS FOR TESTING ************************************************/
// CheckPassword checks if the suplied web access password is valid.
func (c *Client) CheckPassword(user string, password []byte, hotpToken string) error {
	_, err := c.PostJSON(
		c.Endpoint("users", user, "web", "password", "check"),
			checkPasswordReq{
				Password:  string(password),
				HOTPToken: hotpToken,
			})
	return trace.Wrap(err)
}

// -- BEGIN SERVER -- //
srv.POST("/v1/users/:user/web/password/check", httplib.MakeHandler(srv.checkPassword))

type checkPasswordReq struct {
	Password  string `json:"password"`
	HOTPToken string `json:"hotp_token"`
}

func (s *APIServer) checkPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	var req checkPasswordReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}
	user := p[0].Value
	if err := s.a.CheckPassword(user, []byte(req.Password), req.HOTPToken); err != nil {
		return nil, trace.Wrap(err)
	}
	return message(fmt.Sprintf("'%v' user password matches", user)), nil
}

	func (a *AuthWithRoles) CheckPassword(user string, password []byte, hotpToken string) error {
		if err := a.permChecker.HasPermission(a.role, ActionCheckPassword); err != nil {
			return trace.Wrap(err)
		}
		return a.authServer.CheckPassword(user, password, hotpToken)
	}

		// CheckPassword is called on web user or tsh user login
		func (s *IdentityService) CheckPassword(user string, password []byte, hotpToken string) error {
			hash, err := s.GetPasswordHash(user)
			if err != nil {
				return trace.Wrap(err)
			}
			if err = s.IncreaseLoginAttempts(user); err != nil {
				return trace.Wrap(err)
			}
			if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
				return trace.AccessDenied("passwords do not match")
			}
			otp, err := s.GetHOTP(user)
			if err != nil {
				return trace.Wrap(err)
			}
			if !otp.Scan(hotpToken, defaults.HOTPFirstTokensRange) {
				return trace.AccessDenied("bad one time token")
			}
			defer s.ResetLoginAttempts(user)
			if err := s.UpsertHOTP(user, otp); err != nil {
				return trace.Wrap(err)
			}
			return nil
		}
// -- END SERVER -- //
