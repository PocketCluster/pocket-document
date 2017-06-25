hosts.go :: func lookupStaticHost(host string) []string

    dnsclient_unix.go :: func goLookupHostOrder(ctx context.Context, name string, order hostLookupOrder) (addrs []string, err error)

        /*** _FIX_CONF_ ***/ 
        lookup_unix.go :: func lookupHost(ctx context.Context, host string) (addrs []string, err error) {
            order := systemConf().hostLookupOrder(host)
            if order == hostLookupCgo {
                if addrs, err, ok := cgoLookupHost(ctx, host); ok {
                    return addrs, err
                }
                // cgo not available (or netgo); fall back to Go's DNS resolver
                order = hostLookupFilesDNS
            }
            return goLookupHostOrder(ctx, host, order)
        }

            // LookupHost looks up the given host using the local resolver. It returns an array of that host's addresses.
            lookup.go :: func LookupHost(host string) (addrs []string, err error)

        /*** [[DEAD END]] ***/ 
        // goLookupHost is the native Go implementation of LookupHost.
        // Used only if cgoLookupHost refuses to handle the request
        // (that is, only if cgoLookupHost is the stub in cgo_stub.go).
        // Normally we let cgo use the C library resolver instead of
        // depending on our lookup code, so that Go and C get the same
        // answers.
        dnsclient_unix.go :: func goLookupHost(ctx context.Context, name string) (addrs []string, err error)

    /* --- --- */

    /*--- TEST DNE---*/ 
    // lookup entries from /etc/hosts
    dnsclient_unix.go :: func goLookupIPFiles(name string) (addrs []IPAddr)

        dnsclient_unix.go :: func goLookupIPOrder(ctx context.Context, name string, order hostLookupOrder) (addrs []IPAddr, err error)

            /*** _FIX_CONF_ ***/ 
            lookup_unix.go :: func lookupIP(ctx context.Context, host string) (addrs []IPAddr, err error) {
                order := systemConf().hostLookupOrder(host)
                if order == hostLookupCgo {
                    if addrs, err, ok := cgoLookupIP(ctx, host); ok {
                        return addrs, err
                    }
                    // cgo not available (or netgo); fall back to Go's DNS resolver
                    order = hostLookupFilesDNS
                }
                return goLookupIPOrder(ctx, host, order)
            }

                // lookupIPContext looks up a hostname with a context.
                lookup.go :: func lookupIPContext(ctx context.Context, host string) (addrs []IPAddr, err error)

                    // internetAddrList resolves addr, which may be a literal IP
                    // address or a DNS name, and returns a list of internet protocol
                    // family addresses. The result contains at least one address when
                    // error is nil.
                    ipsock.go :: func internetAddrList(ctx context.Context, net, addr string) (addrList, error)

                        // ResolveIPAddr parses addr as an IP address of the form "host" or
                        // "ipv6-host%zone" and resolves the domain name on the network net,
                        // which must be "ip", "ip4" or "ip6".
                        iprawsock.go :: func ResolveIPAddr(net, addr string) (*IPAddr, error)

                        // ResolveTCPAddr parses addr as a TCP address of the form "host:port"
                        // or "[ipv6-host%zone]:port" and resolves a pair of domain name and
                        // port name on the network net, which must be "tcp", "tcp4" or
                        // "tcp6".  A literal address or host name for IPv6 must be enclosed
                        // in square brackets, as in "[::1]:80", "[ipv6-host]:http" or
                        // "[ipv6-host%zone]:80".
                        tcpsock.go :: func ResolveTCPAddr(net, addr string) (*TCPAddr, error)

                        // ResolveUDPAddr parses addr as a UDP address of the form "host:port"
                        // or "[ipv6-host%zone]:port" and resolves a pair of domain name and
                        // port name on the network net, which must be "udp", "udp4" or
                        // "udp6".  A literal address or host name for IPv6 must be enclosed
                        // in square brackets, as in "[::1]:80", "[ipv6-host]:http" or
                        // "[ipv6-host%zone]:80".
                        udpsock.go :: func ResolveUDPAddr(net, addr string) (*UDPAddr, error)

                        // resolverAddrList resolves addr using hint and returns a list of
                        // addresses. The result contains at least one address when error is
                        // nil.
                        dial.go :: func resolveAddrList(ctx context.Context, op, network, addr string, hint Addr) (addrList, error)

                            // DialContext connects to the address on the named network using
                            // the provided context.
                            //
                            // The provided Context must be non-nil. If the context expires before
                            // the connection is complete, an error is returned. Once successfully
                            // connected, any expiration of the context will not affect the
                            // connection.
                            //
                            // See func Dial for a description of the network and address
                            // parameters.
                            dial.go :: func (d *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error)

                                dial.go :: func (d *Dialer) Dial(network, address string) 

                            // Listen announces on the local network address laddr.
                            // The network net must be a stream-oriented network: "tcp", "tcp4",
                            // "tcp6", "unix" or "unixpacket".
                            // For TCP and UDP, the syntax of laddr is "host:port", like "127.0.0.1:8080".
                            // If host is omitted, as in ":8080", Listen listens on all available interfaces
                            // instead of just the interface with the given host address.
                            // See Dial for more details about address syntax.
                            dial.go ::func Listen(net, laddr string) (Listener, error)

                            // ListenPacket announces on the local network address laddr.
                            // The network net must be a packet-oriented network: "udp", "udp4",
                            // "udp6", "ip", "ip4", "ip6" or "unixgram".
                            // For TCP and UDP, the syntax of laddr is "host:port", like "127.0.0.1:8080".
                            // If host is omitted, as in ":8080", ListenPacket listens on all available interfaces
                            // instead of just the interface with the given host address.
                            // See Dial for the syntax of laddr.
                            dial.go ::func ListenPacket(net, laddr string) (PacketConn, error)


                // lookupIPMerge wraps lookupIP, but makes sure that for any given
                // host, only one lookup is in-flight at a time. The returned memory
                // is always owned by the caller.
                lookup.go :: func lookupIPMerge(ctx context.Context, host string) (addrs []IPAddr, err error)

                    // LookupIP looks up host using the local resolver.
                    // It returns an array of that host's IPv4 and IPv6 addresses.
                    lookup.go :: func LookupIP(host string) (ips []IP, err error)

            /*** [[DEAD END]] ***/ 
            // goLookupIP is the native Go implementation of LookupIP. The libc versions are in cgo_*.go.
            dnsclient_unix.go :: func goLookupIP(ctx context.Context, name string) (addrs []IPAddr, err error)


/* --- --- --- */

hosts.go :: func lookupStaticAddr(addr string) []string

    /*--- TEST DNE---*/ 
    // goLookupPTR is the native Go implementation of LookupAddr.
    // Used only if cgoLookupPTR refuses to handle the request (that is,
    // only if cgoLookupPTR is the stub in cgo_stub.go).
    // Normally we let cgo use the C library resolver instead of depending
    // on our lookup code, so that Go and C get the same answers.
    hosts.go :: func goLookupPTR(ctx context.Context, addr string) ([]string, error)

        /*** _FIX_CONF_ ***/ 
        lookup_unix.go :: func lookupAddr(ctx context.Context, addr string) ([]string, error) {
            if systemConf().canUseCgo() {
                if ptrs, err, ok := cgoLookupPTR(ctx, addr); ok {
                    return ptrs, err
                }
            }
            return goLookupPTR(ctx, addr)
        }

            // LookupAddr performs a reverse lookup for the given address, returning a list
            // of names mapping to that address.
            lookup_unix.go :: func LookupAddr(addr string) (names []string, err error)

