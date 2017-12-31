/*

[SSH] conn(192.168.1.105:50901->192.168.1.166:3022, user=root) auth attempt with key ssh-rsa-cert-v01@openssh.com 93:35:ac:c1:f6:f1:30:31:0f:51:39:66:4c:56:50:bc  file="srv/sshserver.go:472" func="srv.(*Server).keyAuth"
TunClient[Node].Dial()                        file="auth/tun.go:669" func="auth.(*TunClient).Dial"
tunClient(Node).authServers: [{192.168.1.244:3025 tcp } {192.168.1.244:3025 tcp }]  file="auth/tun.go:785" func="auth.(*TunClient).getClient"
tunClient.Dial(to=192.168.1.244:3025, attempt=1)  file="auth/tun.go:805" func="auth.(*TunClient).dialAuthServer"
[SSH] successfully authenticated              file="srv/sshserver.go:512" fingerprint="ssh-rsa-cert-v01@openssh.com 93:35:ac:c1:f6:f1:30:31:0f:51:39:66:4c:56:50:bc" func="srv.(*Server).keyAuth" local=192.168.1.166:3022 remote=192.168.1.105:50901 user=root
[SSH:node] new connection 192.168.1.105:50901 -> 192.168.1.166:3022 vesion: SSH-2.0-Go  file="sshutils/server.go:208" func="sshutils.(*Server).handleConnection"

[SSH] ssh.dispatch(req=exec, wantReply=true)  component=node fields=map[local:192.168.1.166:3022 remote:192.168.1.105:50901 login:root teleportUser:root id:16] file="srv/sshserver.go:721" func="srv.(*Server).dispatch"
Exec(cmd=/opt/gopkg/src/github.com/stkim1/pc-node-agent/exec/pocketd/pocketd scp --remote-addr=192.168.1.105:50901 --local-addr=192.168.1.166:3022 -t /pocket/100mb.test) started  component=node fields=map[local:192.168.1.166:3022 remote:192.168.1.105:50901 login:root teleportUser:root id:16] file="srv/exec.go:288" func="srv.(*execResponse).start"
[SSH] ctx.result = {/opt/gopkg/src/github.com/stkim1/pc-node-agent/exec/pocketd/pocketd 2 []}  component=node fields=map[login:root teleportUser:root id:16 local:192.168.1.166:3022 remote:192.168.1.105:50901] file="srv/sshserver.go:704" func="srv.(*Server).handleSessionRequests"
[SSH:node] closed connection                  file="sshutils/server.go:210" func="sshutils.(*Server).handleConnection.func1"

*/

// handleSessionRequests handles out of band session requests once the session channel has been created
// this function's loop handles all the "exec", "subsystem" and "shell" requests.
func (s *Server) handleSessionRequests(sconn *ssh.ServerConn, ch ssh.Channel, in <-chan *ssh.Request) {
    // ctx holds the connection context and keeps track of the associated resources
    ctx := newCtx(s, sconn)
    ctx.isTestStub = s.isTestStub
    ctx.addCloser(ch)
    defer ctx.Close()

    // As SSH conversation progresses, at some point a session will be created and
    // its ID will be added to the environment
    updateContext := func() {
        ssid, found := ctx.getEnv(sshutils.SessionEnvVar)
        if !found {
            return
        }
        findSession := func() (*session, bool) {
            s.reg.Lock()
            defer s.reg.Unlock()
            return s.reg.findSession(rsession.ID(ssid))
        }
        // update ctx with a session ID
        ctx.session, _ = findSession()
        log.Debugf("[SSH] loaded session %v for SSH connection %v", ctx.session, sconn)
    }

    for {
        // update ctx with the session ID:
        if !s.proxyMode {
            updateContext()
        }
        select {
        case creq := <-ctx.subsystemResultC:
            // this means that subsystem has finished executing and
            // want us to close session and the channel
            ctx.Debugf("[SSH] close session request: %v", creq.err)
            return
        case req := <-in:
            if req == nil {
                // this will happen when the client closes/drops the connection
                ctx.Debugf("[SSH] client %v disconnected", sconn.RemoteAddr())
                return
            }
            if err := s.dispatch(ch, req, ctx); err != nil {
                replyError(ch, req, err)
                return
            }
            if req.WantReply {
                req.Reply(true, nil)
            }
        case result := <-ctx.result:
            ctx.Debugf("[SSH] ctx.result = %v", result)
            // pass back stderr output
            if len(result.stderr) != 0 {
                ch.Stderr().Write(result.stderr)
            }
            // this means that exec process has finished and delivered the execution result,
            // we send it back and close the session
            _, err := ch.SendRequest("exit-status", false, ssh.Marshal(struct{ C uint32 }{C: uint32(result.code)}))
            if err != nil {
                ctx.Infof("[SSH] %v failed to send exit status: %v", result.command, err)
            }
            return
        }
    }
}

    // dispatch receives an SSH request for a subsystem and disptaches the request to the
    // appropriate subsystem implementation
    func (s *Server) dispatch(ch ssh.Channel, req *ssh.Request, ctx *ctx) error {
        ctx.Debugf("[SSH] ssh.dispatch(req=%v, wantReply=%v)", req.Type, req.WantReply)

        // if this SSH server is configured to only proxy, we do not support anything other
        // than our own custom "subsystems" and environment manipulation
        if s.proxyMode {
            switch req.Type {
            case "subsystem":
                return s.handleSubsystem(ch, req, ctx)
            case "env":
                // we currently ignore setting any environment variables via SSH for security purposes
                return s.handleEnv(ch, req, ctx)
            default:
                return trace.BadParameter(
                    "proxy doesn't support request type '%v'", req.Type)
            }
        }

        switch req.Type {
        case "exec":
            // exec is a remote execution of a program, does not use PTY
            return s.handleExec(ch, req, ctx)
        case sshutils.PTYReq:
            // SSH client asked to allocate PTY
            return s.handlePTYReq(ch, req, ctx)
        case "shell":
            // SSH client asked to launch shell, we allocate PTY and start shell session
            ctx.exec = &execResponse{ctx: ctx}
            return s.reg.openSession(ch, req, ctx)
        case "env":
            return s.handleEnv(ch, req, ctx)
        case "subsystem":
            // subsystems are SSH subsystems defined in http://tools.ietf.org/html/rfc4254 6.6
            // they are in essence SSH session extensions, allowing to implement new SSH commands
            return s.handleSubsystem(ch, req, ctx)
        case sshutils.WindowChangeReq:
            return s.handleWinChange(ch, req, ctx)
        case "auth-agent-req@openssh.com":
            // This happens when SSH client has agent forwarding enabled, in this case
            // client sends a special request, in return SSH server opens new channel
            // that uses SSH protocol for agent drafted here:
            // https://tools.ietf.org/html/draft-ietf-secsh-agent-02
            // the open ssh proto spec that we implement is here:
            // http://cvsweb.openbsd.org/cgi-bin/cvsweb/src/usr.bin/ssh/PROTOCOL.agent
            return s.handleAgentForward(ch, req, ctx)
        default:
            return trace.BadParameter(
                "proxy doesn't support request type '%v'", req.Type)
        }
    }

        // handleExec is responsible for executing 'exec' SSH requests (i.e. executing
        // a command after making an SSH connection)
        //
        // Note: this also handles 'scp' requests because 'scp' is a subset of "exec"
        func (s *Server) handleExec(ch ssh.Channel, req *ssh.Request, ctx *ctx) error {
            execResponse, err := parseExecRequest(req, ctx)
            if err != nil {
                ctx.Infof("failed to parse exec request: %v", err)
                replyError(ch, req, err)
                return trace.Wrap(err)
            }
            if req.WantReply {
                req.Reply(true, nil)
            }
            // a terminal has been previously allocate for this command.
            // run this inside an interactive session
            if ctx.term != nil {
                return s.reg.openSession(ch, req, ctx)
            }
            // ... otherwise, regular execution:
            result, err := execResponse.start(ch)
            if err != nil {
                ctx.Error(err)
                replyError(ch, req, err)
            }
            if result != nil {
                ctx.Debugf("%v result collected: %v", execResponse, result)
                ctx.sendResult(*result)
            }
            if err != nil {
                return trace.Wrap(err)
            }
            // in case if result is nil and no error, this means that program is
            // running in the background
            go func() {
                result, err := execResponse.wait()
                if err != nil {
                    ctx.Errorf("%v wait failed: %v", execResponse, err)
                }
                if result != nil {
                    ctx.sendResult(*result)
                }
            }()
            return nil
        }

            // parseExecRequest parses SSH exec request
            func parseExecRequest(req *ssh.Request, ctx *ctx) (*execResponse, error) {
                var e execReq
                if err := ssh.Unmarshal(req.Payload, &e); err != nil {
                    return nil, fmt.Errorf("failed to parse exec request, error: %v", err)
                }
                // is this scp request?
                args := strings.Split(e.Command, " ")
                if len(args) > 0 {
                    _, f := filepath.Split(args[0])
                    if f == "scp" {
                        // for 'scp' requests, we'll fork ourselves with scp parameters:
                        teleportBin, err := osext.Executable()
                        if err != nil {
                            return nil, trace.Wrap(err)
                        }
                        e.Command = fmt.Sprintf("%s scp --remote-addr=%s --local-addr=%s %v",
                            teleportBin,
                            ctx.conn.RemoteAddr().String(),
                            ctx.conn.LocalAddr().String(),
                            strings.Join(args[1:], " "))
                    }
                }
                ctx.exec = &execResponse{
                    ctx:     ctx,
                    cmdName: e.Command,
                }
                return ctx.exec, nil
            }

            // start launches the given command returns (nil, nil) if successful. execResult is only used
            // to communicate an error while launching
            func (e *execResponse) start(ch ssh.Channel) (*execResult, error) {
                var err error
                e.cmd, err = prepareCommand(e.ctx)
                if err != nil {
                    return nil, trace.Wrap(err)
                }
                e.cmd.Stderr = ch.Stderr()
                e.cmd.Stdout = ch

                inputWriter, err := e.cmd.StdinPipe()
                if err != nil {
                    return nil, trace.Wrap(err)
                }
                go func() {
                    io.Copy(inputWriter, ch)
                    inputWriter.Close()
                }()

                if err := e.cmd.Start(); err != nil {
                    e.ctx.Warningf("%v start failure err: %v", e, err)
                    return e.collectStatus(e.cmd, trace.ConvertSystemError(err))
                }
                e.ctx.Infof("%v started", e)

                return nil, nil
            }




// scp mode execute command
./pocketd scp --remote-addr=192.168.1.105:50901 --local-addr=192.168.1.166:3022 -t /pocket/100mb.test


// define a hidden 'scp' command (it implements server-side implementation of handling
// 'scp' requests)
scpc.Flag("t", "sink mode (data consumer)").Short('t').Default("false").BoolVar(&scpCommand.Sink)
scpc.Flag("f", "source mode (data producer)").Short('f').Default("false").BoolVar(&scpCommand.Source)
scpc.Flag("v", "verbose mode").Default("false").Short('v').BoolVar(&scpCommand.Verbose)
scpc.Flag("r", "recursive mode").Default("false").Short('r').BoolVar(&scpCommand.Recursive)
scpc.Flag("remote-addr", "address of the remote client").StringVar(&scpCommand.RemoteAddr)
scpc.Flag("local-addr", "local address which accepted the request").StringVar(&scpCommand.LocalAddr)
scpc.Arg("target", "").StringVar(&scpCommand.Target)


// onSCP implements handling of 'scp' requests on the server side. When the teleport SSH daemon
// receives an SSH "scp" request, it launches itself with 'scp' flag under the requested
// user's privileges
//
// This is the entry point of "teleport scp" call (the parent process is the teleport daemon)
func onSCP(cmd *scp.Command) (err error) {
    // get user's home dir (it serves as a default destination)
    cmd.User, err = user.Current()
    if err != nil {
        return trace.Wrap(err)
    }
    // see if the target is absolute. if not, use user's homedir to make
    // it absolute (and if the user doesn't have a homedir, use "/")
    slash := string(filepath.Separator)
    withSlash := strings.HasSuffix(cmd.Target, slash)
    if !filepath.IsAbs(cmd.Target) {
        rootDir := cmd.User.HomeDir
        if !utils.IsDir(rootDir) {
            cmd.Target = slash + cmd.Target
        } else {
            cmd.Target = filepath.Join(rootDir, cmd.Target)
            if withSlash {
                cmd.Target = cmd.Target + slash
            }
        }
    }
    if !cmd.Source && !cmd.Sink {
        return trace.Errorf("remote mode is not supported")
    }
    return trace.Wrap(cmd.Execute(&StdReadWriter{}))
}


