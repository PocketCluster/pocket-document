case start.FullCommand():
	// configuration merge: defaults -> file-based conf -> CLI conf
	if err = config.Configure(&ccf, conf); err != nil {
		utils.FatalError(err)
	}
	if !testRun {
		log.Info(conf.DebugDumpToYAML())
	}
	if ccf.HTTPProfileEndpoint {
		log.Infof("starting http profile endpoint")
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	if !testRun {
		err = onStart(conf)
	}

// Configure merges command line arguments with what's in a configuration file
// with CLI commands taking precedence
func Configure(clf *CommandLineFlags, cfg *service.Config) error {
	// load /etc/teleport.yaml and apply it's values:
	fileConf, err := ReadConfigFile(clf.ConfigFile)
	if err != nil {
		return trace.Wrap(err)
	}
	// if configuration is passed as an environment variable,
	// try to decode it and override the config file
	if clf.ConfigString != "" {
		fileConf, err = ReadFromString(clf.ConfigString)
		if err != nil {
			return trace.Wrap(err)
		}
	}
	if err = ApplyFileConfig(fileConf, cfg); err != nil {
		return trace.Wrap(err)
	}

	// apply --debug flag:
	if clf.Debug {
		cfg.Console = ioutil.Discard
		utils.InitLoggerDebug()
	}

	// apply --roles flag:
	if clf.Roles != "" {
		if err := validateRoles(clf.Roles); err != nil {
			return trace.Wrap(err)
		}
		cfg.SSH.Enabled = strings.Index(clf.Roles, defaults.RoleNode) != -1
		cfg.Auth.Enabled = strings.Index(clf.Roles, defaults.RoleAuthService) != -1
		cfg.Proxy.Enabled = strings.Index(clf.Roles, defaults.RoleProxy) != -1
	}

	// apply --auth-server flag:
	if clf.AuthServerAddr != "" {
		if cfg.Auth.Enabled {
			log.Warnf("not starting the local auth service. --auth-server flag tells to connect to another auth server")
			cfg.Auth.Enabled = false
		}
		addr, err := utils.ParseHostPortAddr(clf.AuthServerAddr, int(defaults.AuthListenPort))
		if err != nil {
			return trace.Wrap(err)
		}
		log.Infof("Using auth server: %v", addr.FullAddress())
		cfg.AuthServers = []utils.NetAddr{*addr}
	}

	// apply --name flag:
	if clf.NodeName != "" {
		cfg.Hostname = clf.NodeName
	}

	// apply --token flag:
	cfg.ApplyToken(clf.AuthToken)

	// apply --listen-ip flag:
	if clf.ListenIP != nil {
		applyListenIP(clf.ListenIP, cfg)
	}

	// --advertise-ip flag
	if clf.AdvertiseIP != nil {
		if err := validateAdvertiseIP(clf.AdvertiseIP); err != nil {
			return trace.Wrap(err)
		}
		cfg.AdvertiseIP = clf.AdvertiseIP
	}

	// apply --labels flag
	if err = parseLabels(clf.Labels, &cfg.SSH); err != nil {
		return trace.Wrap(err)
	}

	// locate web assets if web proxy is enabled
	if cfg.Proxy.Enabled {
		cfg.Proxy.AssetsDir, err = LocateWebAssets()
		if err != nil {
			return trace.Wrap(err)
		}
	}

	// --pid-file:
	if clf.PIDFile != "" {
		cfg.PIDFile = clf.PIDFile
	}

	// auth_servers not configured, but the 'auth' is enabled (auth is on localhost)?
	if len(cfg.AuthServers) == 0 && cfg.Auth.Enabled {
		cfg.AuthServers = append(cfg.AuthServers, cfg.Auth.SSHAddr)
	}

	return nil
}


// onStart is the handler for "start" CLI command
func onStart(config *service.Config) error {
	srv, err := service.NewTeleport(config)
	if err != nil {
		return trace.Wrap(err, "initializing teleport")
	}
	if err := srv.Start(); err != nil {
		return trace.Wrap(err, "starting teleport")
	}

	// create the pid file
	if config.PIDFile != "" {
		f, err := os.OpenFile(config.PIDFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return trace.Wrap(err, "failed to create the PID file")
		}
		fmt.Fprintf(f, "%v", os.Getpid())
		defer f.Close()
	}
	srv.Wait()
	return nil
}

	// NewTeleport takes the daemon configuration, instantiates all required services
	// and starts them under a supervisor, returning the supervisor object
	func NewTeleport(cfg *Config) (*TeleportProcess, error) {
		if err := validateConfig(cfg); err != nil {
			return nil, trace.Wrap(err, "Configuration error")
		}

		// create the data directory if it's missing
		_, err := os.Stat(cfg.DataDir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(cfg.DataDir, os.ModeDir|0700)
			if err != nil {
				return nil, trace.Wrap(err)
			}
		}

		// if there's no host uuid initialized yet, try to read one from the
		// one of the identities
		cfg.HostUUID, err = utils.ReadHostUUID(cfg.DataDir)
		if err != nil {
			if !trace.IsNotFound(err) {
				return nil, trace.Wrap(err)
			}
			if len(cfg.Identities) != 0 {
				cfg.HostUUID = cfg.Identities[0].ID.HostUUID
				log.Infof("[INIT] taking host uuid from first identity: %v", cfg.HostUUID)
			} else {
				cfg.HostUUID = uuid.New()
				log.Infof("[INIT] generating new host UUID: %v", cfg.HostUUID)
			}
			if err := utils.WriteHostUUID(cfg.DataDir, cfg.HostUUID); err != nil {
				return nil, trace.Wrap(err)
			}
		}

		// if user started auth and another service (without providing the auth address for
		// that service, the address of the in-process auth will be used
		if cfg.Auth.Enabled && len(cfg.AuthServers) == 0 {
			cfg.AuthServers = []utils.NetAddr{cfg.Auth.SSHAddr}
		}

		// if user did not provide auth domain name, use this host UUID
		if cfg.Auth.Enabled && cfg.Auth.DomainName == "" {
			cfg.Auth.DomainName = cfg.HostUUID
		}

		// try to login into the auth service:

		// if there are no certificates, use self signed
		process := &TeleportProcess{
			Supervisor: NewSupervisor(),
			Config:     cfg,
		}

		serviceStarted := false

		if cfg.Auth.Enabled {
			if cfg.Keygen == nil {
				cfg.Keygen = native.New()
			}
			if err := process.initAuthService(cfg.Keygen); err != nil {
				return nil, trace.Wrap(err)
			}
			serviceStarted = true
		}

		if cfg.SSH.Enabled {
			if err := process.initSSH(); err != nil {
				return nil, err
			}
			serviceStarted = true
		}

		if cfg.Proxy.Enabled {
			if err := process.initProxy(); err != nil {
				return nil, err
			}
			serviceStarted = true
		}

		if !serviceStarted {
			return nil, trace.Errorf("all services failed to start")
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

					func readToken(token string) (string, error) {
						if !strings.HasPrefix(token, "/") {
							return token, nil
						}
						// treat it as a file
						out, err := ioutil.ReadFile(token)
						if err != nil {
							return "", nil
						}
						return string(out), nil
					}

					type PackedKeys struct {
						Key  []byte `json:"key"`
						Cert []byte `json:"cert"`
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


			// initProxy gets called if teleport runs with 'proxy' role enabled.
			// this means it will do two things:
			//    1. serve a web UI
			//    2. proxy SSH connections to nodes running with 'node' role
			//    3. take care of revse tunnels
			func (process *TeleportProcess) initProxy() error {
				// if no TLS key was provided for the web UI, generate a self signed cert
				if process.Config.Proxy.TLSKey == "" && !process.Config.Proxy.DisableWebUI {
					err := initSelfSignedHTTPSCert(process.Config)
					if err != nil {
						return trace.Wrap(err)
					}
				}

				process.RegisterWithAuthServer(
					process.Config.Token, teleport.RoleProxy,
					ProxyIdentityEvent)

				process.RegisterFunc(func() error {
					eventsC := make(chan Event)
					process.WaitForEvent(ProxyIdentityEvent, eventsC, make(chan struct{}))

					event := <-eventsC
					log.Infof("[SSH] received %v", &event)
					conn, ok := (event.Payload).(*Connector)
					if !ok {
						return trace.BadParameter("unsupported connector type: %T", event.Payload)
					}
					return trace.Wrap(process.initProxyEndpoint(conn))
				})
				return nil
			}

				func (process *TeleportProcess) initProxyEndpoint(conn *Connector) error {
					var (
						askedToExit = true
						err         error
					)
					cfg := process.Config
					proxyLimiter, err := limiter.NewLimiter(cfg.Proxy.Limiter)
					if err != nil {
						return trace.Wrap(err)
					}

					reverseTunnelLimiter, err := limiter.NewLimiter(cfg.Proxy.Limiter)
					if err != nil {
						return trace.Wrap(err)
					}

					tsrv, err := reversetunnel.NewServer(
						cfg.Proxy.ReverseTunnelListenAddr,
						[]ssh.Signer{conn.Identity.KeySigner},
						conn.Client,
						reversetunnel.SetLimiter(reverseTunnelLimiter),
						reversetunnel.DirectSite(conn.Identity.Cert.Extensions[utils.CertExtensionAuthority], conn.Client),
					)
					if err != nil {
						return trace.Wrap(err)
					}

					SSHProxy, err := srv.New(cfg.Proxy.SSHAddr,
						cfg.Hostname,
						[]ssh.Signer{conn.Identity.KeySigner},
						conn.Client,
						cfg.DataDir,
						nil,
						srv.SetLimiter(proxyLimiter),
						srv.SetProxyMode(tsrv),
						srv.SetSessionServer(conn.Client),
						srv.SetAuditLog(conn.Client),
					)
					if err != nil {
						return trace.Wrap(err)
					}

					// Register reverse tunnel agents pool
					agentPool, err := reversetunnel.NewAgentPool(reversetunnel.AgentPoolConfig{
						HostUUID:    conn.Identity.ID.HostUUID,
						Client:      conn.Client,
						HostSigners: []ssh.Signer{conn.Identity.KeySigner},
					})
					if err != nil {
						return trace.Wrap(err)
					}

					// register SSH reverse tunnel server that accepts connections
					// from remote teleport nodes
					process.RegisterFunc(func() error {
						utils.Consolef(cfg.Console, "[PROXY] Reverse tunnel service is starting on %v", cfg.Proxy.ReverseTunnelListenAddr.Addr)
						if err := tsrv.Start(); err != nil {
							utils.Consolef(cfg.Console, "[PROXY] Error: %v", err)
							return trace.Wrap(err)
						}
						// notify parties that we've started reverse tunnel server
						process.BroadcastEvent(Event{Name: ProxyReverseTunnelServerEvent, Payload: tsrv})
						tsrv.Wait()
						if askedToExit {
							log.Infof("[PROXY] Reverse tunnel exited")
						}
						return nil
					})

					// Register web proxy server
					var webListener net.Listener
					if !process.Config.Proxy.DisableWebUI {
						process.RegisterFunc(func() error {
							utils.Consolef(cfg.Console, "[PROXY] Web proxy service is starting on %v", cfg.Proxy.WebAddr.Addr)
							webHandler, err := web.NewHandler(
								web.Config{
									Proxy:       tsrv,
									AssetsDir:   cfg.Proxy.AssetsDir,
									AuthServers: cfg.AuthServers[0],
									DomainName:  cfg.Hostname,
									ProxyClient: conn.Client,
									DisableUI:   cfg.Proxy.DisableWebUI,
								})
							if err != nil {
								utils.Consolef(cfg.Console, "[PROXY] starting the web server: %v", err)
								return trace.Wrap(err)
							}
							defer webHandler.Close()

							proxyLimiter.WrapHandle(webHandler)
							process.BroadcastEvent(Event{Name: ProxyWebServerEvent, Payload: webHandler})

							log.Infof("[PROXY] init TLS listeners")
							webListener, err = utils.ListenTLS(
								cfg.Proxy.WebAddr.Addr,
								cfg.Proxy.TLSCert,
								cfg.Proxy.TLSKey)
							if err != nil {
								return trace.Wrap(err)
							}
							if err = http.Serve(webListener, proxyLimiter); err != nil {
								if askedToExit {
									log.Infof("[PROXY] web server exited")
									return nil
								}
								log.Error(err)
							}
							return nil
						})
					} else {
						log.Infof("[WEB] Web UI is disabled")
					}

					// Register ssh proxy server
					process.RegisterFunc(func() error {
						utils.Consolef(cfg.Console, "[PROXY] SSH proxy service is starting on %v", cfg.Proxy.SSHAddr.Addr)
						if err := SSHProxy.Start(); err != nil {
							if askedToExit {
								log.Infof("[PROXY] SSH proxy exited")
								return nil
							}
							utils.Consolef(cfg.Console, "[PROXY] Error: %v", err)
							return trace.Wrap(err)
						}
						return nil
					})

					process.RegisterFunc(func() error {
						log.Infof("[PROXY] starting reverse tunnel agent pool")
						if err := agentPool.Start(); err != nil {
							log.Fatalf("failed to start: %v", err)
							return trace.Wrap(err)
						}
						agentPool.Wait()
						return nil
					})

					// execute this when process is asked to exit:
					process.onExit(func(payload interface{}) {
						tsrv.Close()
						SSHProxy.Close()
						agentPool.Stop()
						if webListener != nil {
							webListener.Close()
						}
						log.Infof("[PROXY] proxy service exited")
					})
					return nil
				}

func (s *LocalSupervisor) Start() error {
	s.Lock()
	defer s.Unlock()
	s.state = stateStarted

	if len(s.services) == 0 {
		log.Warning("supervisor.Start(): nothing to run")
		return nil
	}

	for _, srv := range s.services {
		s.serve(srv)
	}

	return nil
}

	func (s *LocalSupervisor) serve(srv *Service) {
		// this func will be called _after_ a service stops running:
		removeService := func() {
			s.Lock()
			defer s.Unlock()
			for i, el := range s.services {
				if el == srv {
					s.services = append(s.services[:i], s.services[i+1:]...)
					break
				}
			}
			log.Debugf("[SUPERVISOR] Service %v is done (%v)", *srv, len(s.services))
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer removeService()

			log.Debugf("[SUPERVISOR] Service %v started (%v)", *srv, s.ServiceCount())
			err := (*srv).Serve()
			if err != nil {
				utils.FatalError(err)
			}
		}()
	}
