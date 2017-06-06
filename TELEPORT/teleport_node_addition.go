
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


    // GenerateToken creates a special provisioning token for a new SSH server
    // that is valid for ttl period seconds.
    //
    // This token is used by SSH server to authenticate with Auth server
    // and get signed certificate and private key from the auth server.
    //
    // The token can be used only once.
    func (c *Client) GenerateToken(roles teleport.Roles, ttl time.Duration) (string, error) {
        out, err := c.PostJSON(c.Endpoint("tokens"), generateTokenReq{
            Roles: roles,
            TTL:   ttl,
        })
        if err != nil {
            return "", trace.Wrap(err)
        }
        var token string
        if err := json.Unmarshal(out.Bytes(), &token); err != nil {
            return "", trace.Wrap(err)
        }
        return token, nil
    }