# Teleport Service Flow


## Auth Server Key Generation

- 1. Static Token Addition : *lib/config/configuration.go*

  ```go
  for _, token := range fc.Auth.StaticTokens {
  	roles, tokenValue, err := token.Parse()
  	if err != nil {
  		return trace.Wrap(err)
  	}
  	cfg.Auth.StaticTokens = append(cfg.Auth.StaticTokens, services.ProvisionToken{Token: tokenValue, Roles: roles, Expires: time.Unix(0, 0)})
  }
  ```

- 3. `teleport` token generation : 
	
  > *lib/auth/apiserver.go*
  
  ```go
  func (s *APIServer) generateToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
  	var req *generateTokenReq
  	if err := httplib.ReadJSON(r, &req); err != nil {
  		return nil, trace.Wrap(err)
  	}
  	token, err := s.a.GenerateToken(req.Roles, req.TTL)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	return string(token), nil
  }
  ```

  > *lib/auth/auth_with_roles.go*

  ```go
  func (a *AuthWithRoles) GenerateToken(roles teleport.Roles, ttl time.Duration) (string, error) {
  	if err := a.permChecker.HasPermission(a.role, ActionGenerateToken); err != nil {
  		return "", trace.Wrap(err)
  	}
  	return a.authServer.GenerateToken(roles, ttl)
  }
  ```

  > *lib/auth/auth.go*

  ```go
  func (s *AuthServer) GenerateToken(roles teleport.Roles, ttl time.Duration) (string, error) {
  	for _, role := range roles {
  		if err := role.Check(); err != nil {
  			return "", trace.Wrap(err)
  		}
  	}
  	token, err := utils.CryptoRandomHex(TokenLenBytes)
  	if err != nil {
  		return "", trace.Wrap(err)
  	}
  	if err := s.Provisioner.UpsertToken(token, roles, ttl); err != nil {
  		return "", err
  	}
  	return token, nil
  }
  ```

- 4. `tctl` add user

  ```go
  // Add creates a new sign-up token and prints a token URL to stdout.
  // A user is not created until he visits the sign-up URL and completes the process
  func (u *UserCommand) Add(client *auth.TunClient) error {
  	// if no local logins were specified, default to 'login'
  	if u.allowedLogins == "" {
  		u.allowedLogins = u.login
  	}
  	user := services.TeleportUser{
  		Name:          u.login,
  		AllowedLogins: strings.Split(u.allowedLogins, ","),
  	}
  	if len(u.identities) != 0 {
  		for _, identityVar := range u.identities {
  			vals := strings.SplitN(identityVar, ":", 2)
  			if len(vals) != 2 {
  				return trace.Errorf("bad flag --identity=%v, expected <connector-id>:<email> format", identityVar)
  			}
  			user.OIDCIdentities = append(user.OIDCIdentities, services.OIDCIdentity{ConnectorID: vals[0], Email: vals[1]})
  		}
  	}
  	token, err := client.CreateSignupToken(&user)
  	if err != nil {
  		return err
  	}
  	proxies, err := client.GetProxies()
  	if err != nil {
  		return trace.Wrap(err)
  	}
  	hostname := "teleport-proxy"
  	if len(proxies) == 0 {
  		fmt.Printf("\x1b[1mWARNING\x1b[0m: this Teleport cluster does not have any proxy servers online.\nYou need to start some to be able to login.\n\n")
  	} else {
  		hostname = proxies[0].Hostname
  	}
  
  	// try to auto-suggest the activation link
  	_, proxyPort, err := net.SplitHostPort(u.config.Proxy.WebAddr.Addr)
  	if err != nil {
  		proxyPort = strconv.Itoa(defaults.HTTPListenPort)
  	}
  	url := web.CreateSignupLink(net.JoinHostPort(hostname, proxyPort), token)
  	fmt.Printf("Signup token has been created and is valid for %v seconds. Share this URL with the user:\n%v\n\nNOTE: make sure '%s' is accessible!\n", defaults.MaxSignupTokenTTL.Seconds(), url, hostname)
  	return nil
  }
  ```

## Reverse Tunnel Agent pool

- Reverse tunnen creation point `lib/service/service.go#706`

  ```go
  process.RegisterFunc(func() error {
  	log.Infof("[PROXY] starting reverse tunnel agent pool")
  	if err := agentPool.Start(); err != nil {
  		log.Fatalf("failed to start: %v", err)
  		return trace.Wrap(err)
  	}
  	agentPool.Wait()
  	return nil
  })
  ```

## Generate User Auth

- `tctl` short lived token generation : *tool/tctl/main.go*

  ```go
  // Invite generates a token which can be used to add another SSH node
  // to a cluster
  func (u *NodeCommand) Invite(client *auth.TunClient) error {
  	if u.count < 1 {
  		return trace.BadParameter("count should be > 0, got %v", u.count)
  	}
  	// parse --roles flag
  	roles, err := teleport.ParseRoles(u.roles)
  	if err != nil {
  		return trace.Wrap(err)
  	}
  	var tokens []string
  	for i := 0; i < u.count; i++ {
  		token, err := client.GenerateToken(roles, u.ttl)
  		if err != nil {
  			return trace.Wrap(err)
  		}
  		tokens = append(tokens, token)
  	}

  	authServers, err := client.GetAuthServers()
  	if err != nil {
  		return trace.Wrap(err)
  	}
  	if len(authServers) == 0 {
  		return trace.Errorf("This cluster does not have any auth servers running")
  	}

  	// output format swtich:
  	if u.format == "text" {
  		for _, token := range tokens {
  			fmt.Printf(
  				"The invite token: %v\nRun this on the new node to join the cluster:\n> teleport start --roles=%s --token=%v --auth-server=%v\n\nPlease note:\n",
  				token, strings.ToLower(roles.String()), token, authServers[0].Addr)
  		}
  		fmt.Printf("  - This invitation token will expire in %d minutes\n", int(u.ttl.Minutes()))
  		fmt.Printf("  - %v must be reachable from the new node, see --advertise-ip server flag\n", authServers[0].Addr)
  	} else {
  		out, err := json.Marshal(tokens)
  		if err != nil {
  			return trace.Wrap(err, "failed to marshal tokens")
  		}
  		fmt.Printf(string(out))
  	}
  	return nil
  }
  ```

- Add user

  ```go
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
  ```

- API Server

  ```go
  func (a *AuthWithRoles) GenerateUserCert(key []byte, user string, ttl time.Duration) ([]byte, error) {
  	if err := a.permChecker.HasPermission(a.role, ActionGenerateUserCert); err != nil {
  		return nil, trace.Wrap(err)
  	}
  	return a.authServer.GenerateUserCert(key, user, ttl)
  }
  ```

- AuthServer

  ```go
  // GenerateUserCert generates user certificate, it takes pkey as a signing
  // private key (user certificate authority)
  func (s *AuthServer) GenerateUserCert(
  	key []byte, username string, ttl time.Duration) ([]byte, error) {
  
  	ca, err := s.Trust.GetCertAuthority(services.CertAuthID{
  		Type:       services.UserCA,
  		DomainName: s.DomainName,
  	}, true)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	privateKey, err := ca.FirstSigningKey()
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	user, err := s.GetUser(username)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	return s.Authority.GenerateUserCert(privateKey, key, username, user.GetAllowedLogins(), ttl)
  }
  ```

## Create User via Web

- `lib/web/web.go`

  ```go
  // createNewUser creates new user entry based on the invite token
  //
  // POST /v1/webapi/users
  //
  // {"invite_token": "unique invite token", "pass": "user password", "second_factor_token": "valid second   factor token"}
  //
  // Sucessful response: (session cookie is set)
  //
  // {"type": "bearer", "token": "bearer token", "user": "alex", "expires_in": 20}
  func (m *Handler) createNewUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{},   error) {
  	var req *createNewUserReq
  	if err := httplib.ReadJSON(r, &req); err != nil {
  		return nil, trace.Wrap(err)
  	}
  	sess, err := m.auth.CreateNewUser(req.InviteToken, req.Pass, req.SecondFactorToken)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	ctx, err := m.auth.ValidateSession(sess.Username, sess.ID)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	if err := SetSession(w, sess.Username, sess.ID); err != nil {
  		return nil, trace.Wrap(err)
  	}
  	return NewSessionResponse(ctx)
  }
  ```

  ```go
  func (s *sessionCache) CreateNewUser(token, password, hotpToken string) (*auth.Session, error) {
  	method, err := auth.NewSignupTokenAuth(token)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	clt, err := auth.NewTunClient("web.create-user", s.authServers, "tokenAuth", method)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	defer clt.Close()
  	sess, err := clt.CreateUserWithToken(token, password, hotpToken)
  	return sess, trace.Wrap(err)
  }
  ```
  
  ```go
  // NewTunClient returns an instance of new HTTP client to Auth server API
  // exposed over SSH tunnel, so client  uses SSH credentials to dial and authenticate
  //  - purpose is mostly for debuggin, like "web client" or "reverse tunnel client"
  //  - authServers: list of auth servers in this cluster (they are supposed to be in sync)
  //  - authMethods: how to authenticate (via cert, web passwowrd, etc)
  //  - opts : functional arguments for further extending
  func NewTunClient(purpose string,
  	authServers []utils.NetAddr,
  	user string,
  	authMethods []ssh.AuthMethod,
  	opts ...TunClientOption) (*TunClient, error) {
  	if user == "" {
  		return nil, trace.BadParameter("SSH connection requires a valid username")
  	}
  	tc := &TunClient{
  		purpose:           purpose,
  		user:              user,
  		staticAuthServers: authServers,
  		authMethods:       authMethods,
  		closeC:            make(chan struct{}),
  	}
  	for _, o := range opts {
  		o(tc)
  	}
  	log.Debugf("newTunClient(%s) with auth: %v", purpose, authServers)
  
  	clt, err := NewClient("http://stub:0", tc.Dial)
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	tc.Client = *clt
  
  	// use local information about auth servers if it's available
  	if tc.addrStorage != nil {
  		cachedAuthServers, err := tc.addrStorage.GetAddresses()
  		if err != nil {
  			log.Infof("unable to load the auth server cache: %v", err)
  		} else {
  			tc.setAuthServers(cachedAuthServers)
  		}
  	}
  	return tc, nil
  }
  ```
  
  ```go
  // CreateUserWithToken creates account with provided token and password.
  // Account username and hotp generator are taken from token data.
  // Deletes token after account creation.
  func (c *Client) CreateUserWithToken(token, password, hotpToken string) (*Session, error) {
  	out, err := c.PostJSON(c.Endpoint("signuptokens", "users"), createUserWithTokenReq{
  		Token:     token,
  		Password:  password,
  		HOTPToken: hotpToken,
  	})
  	if err != nil {
  		return nil, trace.Wrap(err)
  	}
  	var sess *Session
  	if err := json.Unmarshal(out.Bytes(), &sess); err != nil {
  		return nil, trace.Wrap(err)
  	}
  	return sess, nil
  }
  ```

### Create user via API

```go
func (s *APIServer) upsertUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	var req *upsertUserReqRaw
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}
	user, err := services.GetUserUnmarshaler()(req.User)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	err = s.a.UpsertUser(user)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return message(fmt.Sprintf("'%v' user upserted", user.GetName())), nil
}
```

```go
func (a *AuthWithRoles) UpsertUser(u services.User) error {
	if err := a.permChecker.HasPermission(a.role, ActionUpsertUser); err != nil {
		return trace.Wrap(err)
	}
	return a.authServer.UpsertUser(u)

}
```
