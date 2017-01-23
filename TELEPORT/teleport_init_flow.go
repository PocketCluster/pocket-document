cfg.Trust = local.NewCAService(cfg.Backend)


// init process

func (process *TeleportProcess) initAuthService(authority auth.Authority) error {
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
	process.setLocalAuth(authServer)
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
