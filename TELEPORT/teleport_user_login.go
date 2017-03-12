/// --- tool/tsh/main.go --- ///
// onLogin logs in with remote proxy and gets signed certificates
func onLogin(cf *CLIConf) {
	tc, err := makeClient(cf, true)
	if err != nil {
		utils.FatalError(err)
	}
	if _, err := tc.Login(); err != nil {
		utils.FatalError(err)
	} else {
		// successful login? update the profile then:
		tc.SaveProfile("")
	}

	if tc.SiteName != "" {
		fmt.Printf("\nYou are now logged into %s as %s\n", tc.SiteName, tc.Username)
	} else {
		fmt.Printf("\nYou are now logged in\n")
	}
}

  /// --- tool/tsh/main.go --- ///
  // makeClient takes the command-line configuration and constructs & returns
  // a fully configured TeleportClient object
  func makeClient(cf *CLIConf, useProfileLogin bool) (tc *client.TeleportClient, err error) {
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

  	// 1: start with the defaults
  	c := client.MakeDefaultConfig()

  	// 2: override with `./tsh` profiles (but only if no proxy is given via the CLI)
  	if cf.Proxy == "" {
  		if err = c.LoadProfile(""); err != nil {
  			fmt.Printf("WARNING: Failed loading tsh profile.\n%v\n", err)
  		}
  	}

  	// 3: override with the CLI flags
  	if cf.Username != "" {
  		c.Username = cf.Username
  	}
  	if cf.Proxy != "" {
  		c.ProxyHostPort = cf.Proxy
  	}
  	if len(fPorts) > 0 {
  		c.LocalForwardPorts = fPorts
  	}
  	if cf.SiteName != "" {
  		c.SiteName = cf.SiteName
  	}
  	// if host logins stored in profiles must be ignored...
  	if !useProfileLogin {
  		c.HostLogin = ""
  	}
  	if hostLogin != "" {
  		c.HostLogin = hostLogin
  	}
  	c.Host = cf.UserHost
  	c.HostPort = int(cf.NodePort)
  	c.Labels = labels
  	c.KeyTTL = time.Minute * time.Duration(cf.MinsToLive)
  	c.InsecureSkipVerify = cf.InsecureSkipVerify
  	c.ConnectorID = cf.ExternalAuth
  	c.Interactive = cf.Interactive
  	return client.NewClient(c)
  }

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
    // MakeKey generates a new unsigned key. It's useless by itself until a
    // trusted CA signs it
    func (tc *TeleportClient) MakeKey() (key *Key, err error) {
    	key = &Key{}
    	keygen := native.New()
    	defer keygen.Close()
    	key.Priv, key.Pub, err = keygen.GenerateKeyPair("")
    	if err != nil {
    		return nil, trace.Wrap(err)
    	}
    	return key, nil
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
      func (c *Config) ProxyWebHostPort() string {
      	return net.JoinHostPort(c.ProxyHost(), strconv.Itoa(c.ProxyWebPort()))
      }

      /// --- lib/client/api.go --- ///
      // loopbackPool reads trusted CAs if it finds it in a predefined location
      // and will work only if target proxy address is loopback
      func loopbackPool(proxyAddr string) *x509.CertPool {
      	if !utils.IsLoopback(proxyAddr) {
      		log.Debugf("not using loopback pool for remote proxy addr: %v", proxyAddr)
      		return nil
      	}
      	log.Debugf("attempting to use loopback pool for local proxy addr: %v", proxyAddr)
      	certPool := x509.NewCertPool()

      	certPath := filepath.Join(defaults.DataDir, defaults.SelfSignedCertPath)
      	pemByte, err := ioutil.ReadFile(certPath)
      	if err != nil {
      		log.Debugf("could not open any path in: %v", certPath)
      		return nil
      	}

      	for {
      		var block *pem.Block
      		block, pemByte = pem.Decode(pemByte)
      		if block == nil {
      			break
      		}
      		cert, err := x509.ParseCertificate(block.Bytes)
      		if err != nil {
      			log.Debugf("could not parse cert in: %v, err: %v", certPath, err)
      			return nil
      		}
      		certPool.AddCert(cert)
      	}
      	log.Debugf("using local pool for loopback proxy: %v, err: %v", certPath, err)
      	return certPool
      }

      /// --- lib/web/sshlogin.go --- ///
      // Ping is used to validate HTTPS endpoing of Teleport proxy. This leads to better
      // user experience: they get connection errors before being asked for passwords
      func Ping(proxyAddr string, insecure bool, pool *x509.CertPool) error {
      	clt, _, err := initClient(proxyAddr, insecure, pool)
      	if err != nil {
      		return trace.Wrap(err)
      	}
      	_, err = clt.Get(clt.Endpoint("webapi"), url.Values{})
      	if err != nil {
      		return trace.Wrap(err)
      	}
      	return nil
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

      /// --- lib/web/sshlogin.go --- ///
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
