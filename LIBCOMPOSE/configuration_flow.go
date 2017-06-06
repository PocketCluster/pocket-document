client.NewPocketCientOption(caCert, tlsCert, tlsKey []byte, host string) (*PocketClientOption, error)

docker.NewPocketProject(pocketContext *PocketContext, parseOptions *config.ParseOptions) (project.APIProject, error)
    auth.NewConfigLookup(configfile *configfile.ConfigFile) *ConfigLookup

    service.NewService(name string, serviceConfig *config.ServiceConfig, context *ctx.Context) *Service

    client.NewPocketFactory(pocketOpts *PocketClientOption) (Factory, error)
        // ref: client.Create()
        client.createPocketClient(pocketOpts *PocketClientOption) (client.APIClient, error)
        client.makeTlsClientConfig(pocketOpts *PocketClientOption) (*tls.Config, error)
            client.pocketCertPool(caCert []byte) (*x509.CertPool, error)
                tlsconfig.SystemCertPool() (*x509.CertPool, error)
                certPool.AppendCertsFromPEM(pemCerts []byte) (ok bool)
            tls.X509KeyPair(certPEMBlock, keyPEMBlock []byte) (Certificate, error)

    project.NewPocketProject(context *Context, runtime RuntimeProject, parseOptions *config.ParseOptions) (*project.Project)

    project.ParseManifest(manifestData []byte) error
        project.Context.openManifest(manifestData []byte) error
            project.Context.parseComposeManifest(manifestData []byte) error
            project.Context.determineProjectFromManifest() error
                project.normalizeName(name string) string

        project.load(file string, bytes []byte) error
            config.Merge(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte, options *ParseOptions) 
                (string, map[string]*ServiceConfig, map[string]*VolumeConfig, map[string]*NetworkConfig, error)

                config.MergeServicesV2(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, datas RawServiceMap, options *ParseOptions) 
                    (map[string]*ServiceConfig, error)

                    config.parseV2(resourceLookup ResourceLookup, environmentLookup EnvironmentLookup, inFile string, serviceData RawService, datas RawServiceMap, options *ParseOptions) 
                        (RawService, error)

                        config.resolveContextV2(inFile string, serviceData RawService) (RawService)

                config.MergeServicesV1(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, datas RawServiceMap, options *ParseOptions) 
                    (map[string]*ServiceConfigV1, error)

    //context.LookupConfig()

err = project.Up(context.Background(), options.Up{})


// --- client.Create() ----
func Create(c Options) (client.APIClient, error) {
    if c.Host == "" {
        if os.Getenv("DOCKER_API_VERSION") == "" {
            os.Setenv("DOCKER_API_VERSION", DefaultAPIVersion)
        }
        client, err := client.NewEnvClient()
            // --- NewEnvClient ---
            func NewEnvClient() (*Client, error) {
                var client *http.Client
                if dockerCertPath := os.Getenv("DOCKER_CERT_PATH"); dockerCertPath != "" {
                    options := tlsconfig.Options{
                        CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
                        CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
                        KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
                        InsecureSkipVerify: os.Getenv("DOCKER_TLS_VERIFY") == "",
                    }
                    tlsc, err := tlsconfig.Client(options)
                        // --- Client --- returns a TLS configuration meant to be used by a client.
                        func Client(options Options) (*tls.Config, error) {
                            tlsConfig := ClientDefault()
                            tlsConfig.InsecureSkipVerify = options.InsecureSkipVerify
                            if !options.InsecureSkipVerify && options.CAFile != "" {
                                CAs, err := certPool(options.CAFile)
                                if err != nil {
                                    return nil, err
                                }
                                tlsConfig.RootCAs = CAs
                            }
                            if options.CertFile != "" || options.KeyFile != "" {
                                tlsCert, err := tls.LoadX509KeyPair(options.CertFile, options.KeyFile)
                                if err != nil {
                                    return nil, fmt.Errorf("Could not load X509 key pair: %v. Make sure the key is not encrypted", err)
                                }
                                tlsConfig.Certificates = []tls.Certificate{tlsCert}
                            }
                            return tlsConfig, nil
                        }
                        // --- Client --- 
                    if err != nil {
                        return nil, err
                    }
                    client = &http.Client{
                        Transport: &http.Transport{
                            TLSClientConfig: tlsc,
                        },
                    }
                }
                host := os.Getenv("DOCKER_HOST")
                if host == "" {
                    host = DefaultDockerHost
                }
                version := os.Getenv("DOCKER_API_VERSION")
                if version == "" {
                    version = DefaultVersion
                }
                cli, err := NewClient(host, version, client, nil)
                if err != nil {
                    return cli, err
                }
                if os.Getenv("DOCKER_API_VERSION") != "" {
                    cli.manualOverride = true
                }
                return cli, nil
            }
            // --- NewEnvClient ---
        if err != nil {
            return nil, err
        }
        return client, nil
    }

    apiVersion := c.APIVersion
    if apiVersion == "" {
        apiVersion = DefaultAPIVersion
    }

    if c.TLSOptions.CAFile == "" {
        c.TLSOptions.CAFile = filepath.Join(dockerCertPath, defaultCaFile)
    }
    if c.TLSOptions.CertFile == "" {
        c.TLSOptions.CertFile = filepath.Join(dockerCertPath, defaultCertFile)
    }
    if c.TLSOptions.KeyFile == "" {
        c.TLSOptions.KeyFile = filepath.Join(dockerCertPath, defaultKeyFile)
    }
    if c.TrustKey == "" {
        c.TrustKey = filepath.Join(homedir.Get(), ".docker", defaultTrustKeyFile)
    }
    if c.TLSVerify {
        c.TLS = true
    }
    if c.TLS {
        c.TLSOptions.InsecureSkipVerify = !c.TLSVerify
    }

    var httpClient *http.Client
    if c.TLS {
        config, err := tlsconfig.Client(c.TLSOptions)
        if err != nil {
            return nil, err
        }
        tr := &http.Transport{
            TLSClientConfig: config,
        }
        proto, addr, _, err := client.ParseHost(c.Host)
        if err != nil {
            return nil, err
        }

        if err := sockets.ConfigureTransport(tr, proto, addr); err != nil {
            return nil, err
        }

        httpClient = &http.Client{
            Transport: tr,
        }
    }

    customHeaders := map[string]string{}
    customHeaders["User-Agent"] = fmt.Sprintf("Libcompose-Client/%s (%s)", version.VERSION, runtime.GOOS)

    client, err := client.NewClient(c.Host, apiVersion, httpClient, customHeaders)
    if err != nil {
        return nil, err
    }
    return client, nil
}

