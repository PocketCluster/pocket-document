dnsclient_unix.go :: func LookupCNAME(name string) (cname string, err error)
    func lookupCNAME(ctx context.Context, name string) (string, error)
        func goLookupCNAME(ctx context.Context, name string) (cname string, err error)

dnsclient_unix.go :: func LookupSRV(service, proto, name string) (cname string, addrs []*SRV, err error)
    func lookupSRV(ctx context.Context, service, proto, name string) (string, []*SRV, error)

dnsclient_unix.go :: func LookupMX(name string) (mxs []*MX, err error)
    func lookupMX(ctx context.Context, name string) ([]*MX, error)

dnsclient_unix.go :: func LookupNS(name string) (nss []*NS, err error)
    func lookupNS(ctx context.Context, name string) ([]*NS, error)

dnsclient_unix.go :: func LookupTXT(name string) (txts []string, err error)
    func lookupTXT(ctx context.Context, name string) ([]string, error)

dnsclient_unix.go :: func LookupAddr(addr string) (names []string, err error)
    func lookupAddr(ctx context.Context, addr string) ([]string, error)

    /* - - - */

    dnsclient_unix.go :: func lookup(ctx context.Context, name string, qtype uint16) (cname string, rrs []dnsRR, err error) {
        if !isDomainName(name) {
            return "", nil, &DNSError{Err: "invalid domain name", Name: name}
        }
        resolvConf.tryUpdate("/etc/resolv.conf")
        resolvConf.mu.RLock()
        conf := resolvConf.dnsConfig
        resolvConf.mu.RUnlock()
        for _, fqdn := range conf.nameList(name) {
            cname, rrs, err = tryOneName(ctx, conf, fqdn, qtype)
            if err == nil {
                break
            }
        }
        if err, ok := err.(*DNSError); ok {
            // Show original name passed to lookup, not suffixed one.
            // In general we might have tried many suffixes; showing
            // just one is misleading. See also golang.org/issue/6324.
            err.Name = name
        }
        return
    }

        // tryUpdate tries to update conf with the named resolv.conf file.
        // The name variable only exists for testing. It is otherwise always
        // "/etc/resolv.conf".
        dnsclient_unix.go :: func (conf *resolverConfig) tryUpdate(name string) {
            conf.initOnce.Do(conf.init)

            // Ensure only one update at a time checks resolv.conf.
            if !conf.tryAcquireSema() {
                return
            }
            defer conf.releaseSema()

            now := time.Now()
            if conf.lastChecked.After(now.Add(-5 * time.Second)) {
                return
            }
            conf.lastChecked = now

            var mtime time.Time
            if fi, err := os.Stat(name); err == nil {
                mtime = fi.ModTime()
            }
            if mtime.Equal(conf.dnsConfig.mtime) {
                return
            }

            dnsConf := dnsReadConfig(name)
            conf.mu.Lock()
            conf.dnsConfig = dnsConf
            conf.mu.Unlock()
        }

            // init initializes conf and is only called via conf.initOnce.
            dnsclient_unix.go :: func (conf *resolverConfig) init() {
                // Set dnsConfig and lastChecked so we don't parse
                // resolv.conf twice the first time.
                conf.dnsConfig = systemConf().resolv
                if conf.dnsConfig == nil {
                    conf.dnsConfig = dnsReadConfig("/etc/resolv.conf")
                }
                conf.lastChecked = time.Now()

                // Prepare ch so that only one update of resolverConfig may
                // run at once.
                conf.ch = make(chan struct{}, 1)
            }

                // See resolv.conf(5) on a Linux machine.
                dnsconfig_unix.go :: func dnsReadConfig(filename string) *dnsConfig {
                    conf := &dnsConfig{
                        ndots:    1,
                        timeout:  5 * time.Second,
                        attempts: 2,
                    }
                    file, err := open(filename)
                    if err != nil {
                        conf.servers = defaultNS
                        conf.search = dnsDefaultSearch()
                        conf.err = err
                        return conf
                    }
                    defer file.close()
                    if fi, err := file.file.Stat(); err == nil {
                        conf.mtime = fi.ModTime()
                    } else {
                        conf.servers = defaultNS
                        conf.search = dnsDefaultSearch()
                        conf.err = err
                        return conf
                    }
                    for line, ok := file.readLine(); ok; line, ok = file.readLine() {
                        if len(line) > 0 && (line[0] == ';' || line[0] == '#') {
                            // comment.
                            continue
                        }
                        f := getFields(line)
                        if len(f) < 1 {
                            continue
                        }
                        switch f[0] {
                        case "nameserver": // add one name server
                            if len(f) > 1 && len(conf.servers) < 3 { // small, but the standard limit
                                // One more check: make sure server name is
                                // just an IP address. Otherwise we need DNS
                                // to look it up.
                                if parseIPv4(f[1]) != nil {
                                    conf.servers = append(conf.servers, JoinHostPort(f[1], "53"))
                                } else if ip, _ := parseIPv6(f[1], true); ip != nil {
                                    conf.servers = append(conf.servers, JoinHostPort(f[1], "53"))
                                }
                            }

                        case "domain": // set search path to just this domain
                            if len(f) > 1 {
                                conf.search = []string{ensureRooted(f[1])}
                            }

                        case "search": // set search path to given servers
                            conf.search = make([]string, len(f)-1)
                            for i := 0; i < len(conf.search); i++ {
                                conf.search[i] = ensureRooted(f[i+1])
                            }

                        case "options": // magic options
                            for _, s := range f[1:] {
                                switch {
                                case hasPrefix(s, "ndots:"):
                                    n, _, _ := dtoi(s, 6)
                                    if n < 1 {
                                        n = 1
                                    }
                                    conf.ndots = n
                                case hasPrefix(s, "timeout:"):
                                    n, _, _ := dtoi(s, 8)
                                    if n < 1 {
                                        n = 1
                                    }
                                    conf.timeout = time.Duration(n) * time.Second
                                case hasPrefix(s, "attempts:"):
                                    n, _, _ := dtoi(s, 9)
                                    if n < 1 {
                                        n = 1
                                    }
                                    conf.attempts = n
                                case s == "rotate":
                                    conf.rotate = true
                                default:
                                    conf.unknownOpt = true
                                }
                            }

                        case "lookup":
                            // OpenBSD option:
                            // http://www.openbsd.org/cgi-bin/man.cgi/OpenBSD-current/man5/resolv.conf.5
                            // "the legal space-separated values are: bind, file, yp"
                            conf.lookup = f[1:]

                        default:
                            conf.unknownOpt = true
                        }
                    }
                    if len(conf.servers) == 0 {
                        conf.servers = defaultNS
                    }
                    if len(conf.search) == 0 {
                        conf.search = dnsDefaultSearch()
                    }
                    return conf
                }

                dnsconfig_unix.go :: var (
                    defaultNS   = []string{"127.0.0.1:53", "[::1]:53"}
                    getHostname = os.Hostname // variable for testing
                )


            // Do a lookup for a single name, which must be rooted
            // (otherwise answer will not find the answers).
            func tryOneName(ctx context.Context, cfg *dnsConfig, name string, qtype uint16) (string, []dnsRR, error) {
            	if len(cfg.servers) == 0 {
            		return "", nil, &DNSError{Err: "no DNS servers", Name: name}
            	}

            	var lastErr error
            	for i := 0; i < cfg.attempts; i++ {
            		for _, server := range cfg.servers {
            			msg, err := exchange(ctx, server, name, qtype, cfg.timeout)
            			if err != nil {
            				lastErr = &DNSError{
            					Err:    err.Error(),
            					Name:   name,
            					Server: server,
            				}
            				if nerr, ok := err.(Error); ok && nerr.Timeout() {
            					lastErr.(*DNSError).IsTimeout = true
            				}
            				continue
            			}
            			// libresolv continues to the next server when it receives
            			// an invalid referral response. See golang.org/issue/15434.
            			if msg.rcode == dnsRcodeSuccess && !msg.authoritative && !msg.recursion_available && len(msg.answer) == 0 && len(msg.extra) == 0 {
            				lastErr = &DNSError{Err: "lame referral", Name: name, Server: server}
            				continue
            			}
            			cname, rrs, err := answer(name, server, msg, qtype)
            			// If answer errored for rcodes dnsRcodeSuccess or dnsRcodeNameError,
            			// it means the response in msg was not useful and trying another
            			// server probably won't help. Return now in those cases.
            			// TODO: indicate this in a more obvious way, such as a field on DNSError?
            			if err == nil || msg.rcode == dnsRcodeSuccess || msg.rcode == dnsRcodeNameError {
            				return cname, rrs, err
            			}
            			lastErr = err
            		}
            	}
            	return "", nil, lastErr
            }
