// onListNodes executes 'tsh ls' command
func onListNodes(cf *CLIConf) {
    tc, err := makeClient(cf, true)
    if err != nil {
        utils.FatalError(err)
    }
    servers, err := tc.ListNodes(context.TODO())
    if err != nil {
        utils.FatalError(err)
    }
    nodesView := func(nodes []services.Server) string {
        t := goterm.NewTable(0, 10, 5, ' ', 0)
        printHeader(t, []string{"Node Name", "Node ID", "Address", "Labels"})
        if len(nodes) == 0 {
            return t.String()
        }
        for _, n := range nodes {
            fmt.Fprintf(t, "%v\t%v\t%v\t%v\n", n.Hostname, n.ID, n.Addr, n.LabelsString())
        }
        return t.String()
    }
    fmt.Printf(nodesView(servers))
}

    type NodeWithSessions struct {
        Node     services.Server   `json:"node"`
        Sessions []session.Session `json:"sessions"`
    }

    type getSiteNodesResponse struct {
        Nodes []nodeWithSessions `json:"nodes"`
    }

    func (m *Handler) getSiteNodes(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *SessionContext, site reversetunnel.RemoteSite) (interface{}, error) {
        clt, err := site.GetClient()
        if err != nil {
            return nil, trace.Wrap(err)
        }
        servers, err := clt.GetNodes()
        if err != nil {
            return nil, trace.Wrap(err)
        }
        sessions, err := clt.GetSessions()
        if err != nil {
            return nil, trace.Wrap(err)
        }
        nodeMap := make(map[string]*nodeWithSessions, len(servers))
        for i := range servers {
            nodeMap[servers[i].ID] = &nodeWithSessions{Node: servers[i]}
        }
        for i := range sessions {
            sess := sessions[i]
            for _, p := range sess.Parties {
                if _, ok := nodeMap[p.ServerID]; ok {
                    nodeMap[p.ServerID].Sessions = append(nodeMap[p.ServerID].Sessions, sess)
                }
            }
        }
        nodes := make([]nodeWithSessions, 0, len(nodeMap))
        for key := range nodeMap {
            nodes = append(nodes, *nodeMap[key])
        }
        return getSiteNodesResponse{
            Nodes: nodes,
        }, nil
    }

        type AuthServer struct {
        }

            // GetNodes returns a list of registered servers
            func (s *PresenceService) GetNodes() ([]services.Server, error) {
                return s.getServers(nodesPrefix)
            }

        sessionService, err := session.New(bkEnd)
        type server struct {
            bk               backend.JSONCodec
            clock            timetools.TimeProvider
            activeSessionTTL time.Duration
        }

            // GetSessions returns a list of active sessions. Returns an empty slice
            // if no sessions are active
            func (s *server) GetSessions() ([]Session, error) {
                bucket := activeBucket()
                out := make(Sessions, 0)

                keys, err := s.bk.GetKeys(bucket)
                if err != nil {
                    log.Error(err)
                    return nil, err
                }
                for i, sid := range keys {
                    if i > MaxSessionSliceLength {
                        break
                    }
                    se, err := s.GetSession(ID(sid))
                    if trace.IsNotFound(err) {
                        continue
                    }
                    out = append(out, *se)
                }
                sort.Stable(out)
                return out, nil
            }

