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

  func NewSupervisor() Supervisor {
  	srv := &LocalSupervisor{
  		services:     []*Service{},
  		wg:           &sync.WaitGroup{},
  		events:       map[string]Event{},
  		eventsC:      make(chan Event, 100),
  		eventWaiters: make(map[string][]*waiter),
  		closer:       utils.NewCloseBroadcaster(),
  	}
  	go srv.fanOut()
  	return srv
  }

  func (s *LocalSupervisor) WaitForEvent(name string, eventC chan Event, cancelC chan struct{}) {
  	s.Lock()
  	defer s.Unlock()

  	waiter := &waiter{eventC: eventC, cancelC: cancelC}
  	event, ok := s.events[name]
  	if ok {
  		go s.notifyWaiter(waiter, event)
  		return
  	}
  	s.eventWaiters[name] = append(s.eventWaiters[name], waiter)
  }

  func (p *PocketNodeProcess) initSSH() error {
      eventsC := make(chan service.Event)

      // register & generate a signed ssh pub/prv key set
      p.RegisterWithAuthServer(p.Config.Token, teleport.RoleNode, service.SSHIdentityEvent)
      p.WaitForEvent(service.SSHIdentityEvent, eventsC, make(chan struct{}))

      // generates a signed certificate & private key for docker/registry
      p.RequestSignedCertificateWithAuthServer(p.Config.Token, teleport.RoleNode, SignedCertificateIssuedEvent)
      p.WaitForEvent(SignedCertificateIssuedEvent, eventsC, make(chan struct{}))

      var s *srv.Server
      p.RegisterFunc(func() error {
          event := <-eventsC
          log.Infof("[SSH] received %v", &event)
          conn, ok := (event.Payload).(*service.Connector)
          if !ok {
              return trace.BadParameter("unsupported connector type: %T", event.Payload)
          }

          cfg := p.Config

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
      p.onExit(func(payload interface{}) {
          s.Close()
      })
      return nil
  }

    // New returns an unstarted server
    func New(addr utils.NetAddr,
    	hostname string,
    	signers []ssh.Signer,
    	authService auth.AccessPoint,
    	dataDir string,
    	advertiseIP net.IP,
    	options ...ServerOption) (*Server, error) {

    	// read the host UUID:
    	uuid, err := utils.ReadOrMakeHostUUID(dataDir)
    	if err != nil {
    		return nil, trace.Wrap(err)
    	}

    	s := &Server{
    		addr:        addr,
    		authService: authService,
    		resolver:    &backendResolver{authService: authService},
    		hostname:    hostname,
    		labelsMutex: &sync.Mutex{},
    		advertiseIP: advertiseIP,
    		uuid:        uuid,
    		closer:      utils.NewCloseBroadcaster(),
    	}
    	s.limiter, err = limiter.NewLimiter(limiter.LimiterConfig{})
    	if err != nil {
    		return nil, trace.Wrap(err)
    	}
    	s.certChecker = ssh.CertChecker{IsAuthority: s.isAuthority}

    	for _, o := range options {
    		if err := o(s); err != nil {
    			return nil, trace.Wrap(err)
    		}
    	}

    	var component string
    	if s.proxyMode {
    		component = teleport.ComponentProxy
    	} else {
    		component = teleport.ComponentNode
    	}

    	s.reg = newSessionRegistry(s)
    	srv, err := sshutils.NewServer(
    		component,
    		addr, s, signers,
    		sshutils.AuthMethods{PublicKey: s.keyAuth},
    		sshutils.SetLimiter(s.limiter),
    		sshutils.SetRequestHandler(s))
    	if err != nil {
    		return nil, trace.Wrap(err)
    	}
    	s.srv = srv
    	return s, nil
    }
