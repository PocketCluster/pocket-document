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

	// initAuthService can be called to initialize auth server service
	func (p *PocketCoreProcess) initAuthService(authority auth.Authority) error {
	    var (
	        askedToExit = false
	        err         error
	    )
	    cfg := p.Config
	    // Initialize the storage back-ends for keys, events and records
	    b, err := p.initAuthStorage()
	    if err != nil {
	        return trace.Wrap(err)
	    }

	    // create the audit log, which will be consuming (and recording) all events
	    // and record sessions
	    var auditLog events.IAuditLog
	    if cfg.Auth.NoAudit {
	        auditLog = &events.DiscardAuditLog{}
	        log.Warn("the audit and session recording are turned off")
	    } else {
	        auditLog, err = events.NewAuditLog(filepath.Join(cfg.DataDir, "log"))
	        if err != nil {
	            return trace.Wrap(err)
	        }
	    }

	    // first, create the AuthServer
	    authServer, identity, err := auth.Init(auth.InitConfig{
	        Backend:         b,
	        Authority:       authority,
	        DomainName:      cfg.Auth.DomainName,
	        AuthServiceName: cfg.Hostname,
	        DataDir:         cfg.DataDir,
	        HostUUID:        cfg.HostUUID,
	        Authorities:     cfg.Auth.Authorities,
	        ReverseTunnels:  cfg.ReverseTunnels,
	        OIDCConnectors:  cfg.OIDCConnectors,
	        Trust:           cfg.Trust,
	        Lock:            cfg.Lock,
	        Presence:        cfg.Presence,
	        Provisioner:     cfg.Provisioner,
	        Identity:        cfg.Identity,
	        StaticTokens:    cfg.Auth.StaticTokens,
	    }, cfg.SeedConfig)
	    if err != nil {
	        return trace.Wrap(err)
	    }
	    p.setLocalAuth(authServer)

	    // second, create the API Server: it's actually a collection of API servers,
	    // each serving requests for a "role" which is assigned to every connected
	    // client based on their certificate (user, server, admin, etc)
	    sessionService, err := session.New(b)
	    if err != nil {
	        return trace.Wrap(err)
	    }
	    apiConf := &auth.APIConfig{
	        AuthServer:        authServer,
	        SessionService:    sessionService,
	        PermissionChecker: auth.NewStandardPermissions(),
	        AuditLog:          auditLog,
	    }

	    limiter, err := limiter.NewLimiter(cfg.Auth.Limiter)
	    if err != nil {
	        return trace.Wrap(err)
	    }

	    // Register an SSH endpoint which is used to create an SSH tunnel to send HTTP
	    // requests to the Auth API
	    var authTunnel *auth.PocketAuthTunnel
	    p.RegisterFunc(func() error {
	        utils.Consolef(cfg.Console, "[AUTH]  Auth service is starting on %v", cfg.Auth.SSHAddr.Addr)
	        authTunnel, err = auth.NewPocketTunnel(
	            cfg.Auth.SSHAddr,
	            identity.KeySigner,
	            p.Config.CaSigner,
	            apiConf,
	            auth.SetLimiter(limiter),
	        )
	        if err != nil {
	            utils.Consolef(cfg.Console, "[AUTH] Error: %v", err)
	            return trace.Wrap(err)
	        }
	        if err := authTunnel.Start(); err != nil {
	            if askedToExit {
	                log.Infof("[AUTH] Auth Tunnel exited")
	                return nil
	            }
	            utils.Consolef(cfg.Console, "[AUTH] Error: %v", err)
	            return trace.Wrap(err)
	        }
	        return nil
	    })

	    p.RegisterFunc(func() error {
	        // Heart beat auth server presence, this is not the best place for this
	        // logic, consolidate it into auth package later
	        connector, err := p.connectToAuthService(teleport.RoleAdmin)
	        if err != nil {
	            return trace.Wrap(err)
	        }
	        // External integrations rely on this event:
	        p.BroadcastEvent(service.Event{Name: service.AuthIdentityEvent, Payload: connector})
	        p.onExit(func(payload interface{}) {
	            connector.Client.Close()
	        })
	        return nil
	    })

	    p.RegisterFunc(func() error {
	        srv := services.Server{
	            ID:       p.Config.HostUUID,
	            Addr:     cfg.Auth.SSHAddr.Addr,
	            Hostname: p.Config.Hostname,
	        }
	        host, port, err := net.SplitHostPort(srv.Addr)
	        // advertise-ip is explicitly set:
	        if p.Config.AdvertiseIP != nil {
	            if err != nil {
	                return trace.Wrap(err)
	            }
	            srv.Addr = fmt.Sprintf("%v:%v", p.Config.AdvertiseIP.String(), port)
	        } else {
	            // advertise-ip is not set, while the CA is listening on 0.0.0.0? lets try
	            // to guess the 'advertise ip' then:
	            if net.ParseIP(host).IsUnspecified() {
	                ip, err := utils.GuessHostIP()
	                if err != nil {
	                    log.Warn(err)
	                } else {
	                    srv.Addr = net.JoinHostPort(ip.String(), port)
	                }
	            }
	            log.Warnf("advertise_ip is not set for this auth server!!! Trying to guess the IP this server can be reached at: %v", srv.Addr)
	        }
	        // immediately register, and then keep repeating in a loop:
	        for !askedToExit {
	            err := authServer.UpsertAuthServer(srv, defaults.ServerHeartbeatTTL)
	            if err != nil {
	                log.Warningf("failed to announce presence: %v", err)
	            }
	            sleepTime := defaults.ServerHeartbeatTTL/2 + utils.RandomDuration(defaults.ServerHeartbeatTTL/10)
	            time.Sleep(sleepTime)
	        }
	        log.Infof("[AUTH] heartbeat to other auth servers exited")
	        return nil
	    })

	    // execute this when process is asked to exit:
	    p.onExit(func(payload interface{}) {
	        askedToExit = true
	        authTunnel.Close()
	        log.Infof("[AUTH] auth service exited")
	    })
	    return nil
	}

		// Init instantiates and configures an instance of AuthServer
		func Init(cfg InitConfig, seedConfig bool) (*AuthServer, *Identity, error) {
			if cfg.DataDir == "" {
				return nil, nil, trace.BadParameter("DataDir: data dir can not be empty")
			}
			if cfg.HostUUID == "" {
				return nil, nil, trace.BadParameter("HostUUID: host UUID can not be empty")
			}

			lockService := local.NewLockService(cfg.Backend)
			err := lockService.AcquireLock(cfg.DomainName, 60*time.Second)
			if err != nil {
				return nil, nil, err
			}
			defer lockService.ReleaseLock(cfg.DomainName)

			// check that user CA and host CA are present and set the certs if needed
			asrv := NewAuthServer(&cfg)

			// we determine if it's the first start by checking if the CA's are set
			firstStart, err := isFirstStart(asrv, cfg)
			if err != nil {
				return nil, nil, trace.Wrap(err)
			}

			// we skip certain configuration if 'seed_config' is set to true
			// and this is NOT the first time teleport starts on this machine
			skipConfig := seedConfig && !firstStart

			// add trusted authorities from the configuration into the trust backend:
			keepMap := make(map[string]int, 0)
			if !skipConfig {
				for _, ca := range cfg.Authorities {
					if err := asrv.Trust.UpsertCertAuthority(ca, backend.Forever); err != nil {
						return nil, nil, trace.Wrap(err)
					}
					keepMap[ca.DomainName] = 1
				}
			}
			// delete trusted authorities from the trust back-end if they're not
			// in the configuration:
			if !seedConfig {
				hostCAs, err := asrv.Trust.GetCertAuthorities(services.HostCA, false)
				if err != nil {
					return nil, nil, trace.Wrap(err)
				}
				userCAs, err := asrv.Trust.GetCertAuthorities(services.UserCA, false)
				if err != nil {
					return nil, nil, trace.Wrap(err)
				}
				for _, ca := range append(hostCAs, userCAs...) {
					_, configured := keepMap[ca.DomainName]
					if ca.DomainName != cfg.DomainName && !configured {
						if err = asrv.Trust.DeleteCertAuthority(*ca.ID()); err != nil {
							return nil, nil, trace.Wrap(err)
						}
						log.Infof("removed old trusted CA: '%s'", ca.DomainName)
					}
				}
			}
			// this block will generate user CA authority on first start if it's
			// not currently present, it will also use optional passed user ca keypair
			// that can be supplied in configuration
			if _, err := asrv.GetCertAuthority(services.CertAuthID{DomainName: cfg.DomainName, Type: services.HostCA}, false); err != nil {
				if !trace.IsNotFound(err) {
					return nil, nil, trace.Wrap(err)
				}
				log.Infof("FIRST START: Generating host CA on first start")
				priv, pub, err := asrv.GenerateKeyPair("")
				if err != nil {
					return nil, nil, trace.Wrap(err)
				}
				hostCA := services.CertAuthority{
					DomainName:   cfg.DomainName,
					Type:         services.HostCA,
					SigningKeys:  [][]byte{priv},
					CheckingKeys: [][]byte{pub},
				}
				if err := asrv.Trust.UpsertCertAuthority(hostCA, backend.Forever); err != nil {
					return nil, nil, trace.Wrap(err)
				}
			}
			// this block will generate user CA authority on first start if it's
			// not currently present, it will also use optional passed user ca keypair
			// that can be supplied in configuration
			if _, err := asrv.GetCertAuthority(services.CertAuthID{DomainName: cfg.DomainName, Type: services.UserCA}, false); err != nil {
				if !trace.IsNotFound(err) {
					return nil, nil, trace.Wrap(err)
				}

				log.Infof("FIRST START: Generating user CA on first start")
				priv, pub, err := asrv.GenerateKeyPair("")
				if err != nil {
					return nil, nil, trace.Wrap(err)
				}
				userCA := services.CertAuthority{
					DomainName:   cfg.DomainName,
					Type:         services.UserCA,
					SigningKeys:  [][]byte{priv},
					CheckingKeys: [][]byte{pub},
				}
				if err := asrv.Trust.UpsertCertAuthority(userCA, backend.Forever); err != nil {
					return nil, nil, trace.Wrap(err)
				}
			}
			// add reverse runnels from the configuration into the backend
			keepMap = make(map[string]int, 0)
			if !skipConfig {
				log.Infof("Initializing reverse tunnels")
				for _, tunnel := range cfg.ReverseTunnels {
					if err := asrv.UpsertReverseTunnel(tunnel, 0); err != nil {
						return nil, nil, trace.Wrap(err)
					}
					keepMap[tunnel.DomainName] = 1
				}
			}
			// remove the reverse tunnels from the backend if they're not
			// present in the configuration
			if !seedConfig {
				tunnels, err := asrv.GetReverseTunnels()
				if err != nil {
					return nil, nil, trace.Wrap(err)
				}
				for _, tunnel := range tunnels {
					_, configured := keepMap[tunnel.DomainName]
					if !configured {
						if err = asrv.DeleteReverseTunnel(tunnel.DomainName); err != nil {
							return nil, nil, trace.Wrap(err)
						}
						log.Infof("removed reverse tunnel: '%s'", tunnel.DomainName)
					}
				}
			}
			// add OIDC connectors to the back-end:
			keepMap = make(map[string]int, 0)
			if !skipConfig {
				log.Infof("Initializing OIDC connectors")
				for _, connector := range cfg.OIDCConnectors {
					if err := asrv.UpsertOIDCConnector(connector, 0); err != nil {
						return nil, nil, trace.Wrap(err)
					}
					log.Infof("created ODIC connector '%s'", connector.ID)
					keepMap[connector.ID] = 1
				}
			}
			// remove OIDC connectors from the backend if they're not
			// present in the configuration
			if !seedConfig {
				connectors, _ := asrv.GetOIDCConnectors(false)
				for _, connector := range connectors {
					_, configured := keepMap[connector.ID]
					if !configured {
						if err = asrv.DeleteOIDCConnector(connector.ID); err != nil {
							return nil, nil, trace.Wrap(err)
						}
						log.Infof("removed OIDC connector '%s'", connector.ID)
					}
				}
			}

			identity, err := initKeys(asrv, cfg.DataDir,
				IdentityID{HostUUID: cfg.HostUUID, Role: teleport.RoleAdmin})
			if err != nil {
				return nil, nil, trace.Wrap(err)
			}
			return asrv, identity, nil
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
