# CA Refresh for Private TLS Conn on Docker / Registry

When new TLS certificate is generated and instated to a node, the TLS cert is not going to be acknowledged by simply placing CA certificate at `/etc/docker/certs.d/<registry>`. We have to refresh system CA cert pool for docker engine to recognize the new TLS.

1. Download the TLS cert. (This must be done automatically by pocketd)
2. Copy `pocketcluster.crt` at `/etc/docker/certs.d/pc-master/`
3. Copy `pocketcluster.crt` at `/usr/local/share/ca-certificates`
4. Run `update-ca-certificates`
5. Restart `docker` service


```sh
# 2
cp /etc/docker/ca-cert.pub /usr/local/share/ca-certificates/pocketcluster.crt

# 3
mkdir -p /etc/docker/certs.d/pc-master/ 
cp /etc/docker/ca-cert.pub /etc/docker/certs.d/pc-master/pocketcluster.crt

# 4
update-ca-certificates 

# 5
service docker restart
```

## Remarks

- `update-ca-certificates` is just a script. We might be able to include it in `pocketd`. Here's another method from [RancherOS](https://docs.rancher.com/rancher/v1.1/en/environments/registries/#self-signed-certificates).

  ```sh
  # Download the certificate from the domain
  $ openssl s_client -showcerts -connect ${DOMAIN}:${PORT} </dev/null 2>/dev/null|openssl x509 -outform PEM > ca.crt

  # Copy the certificate to the appropriate directories
  $ sudo cp ca.crt /etc/docker/certs.d/${DOMAIN}/ca.crt

  # Append the certificate to a file
  $ cat ca.crt | sudo tee -a /etc/ssl/certs/ca-certificates.crt

  # Restart the docker service to have the changes take affect
  $ sudo service docker restart
  ```



- we might be able to create a new CA certificate pool and add `pocketcluster.crt` there to isolate docker engine. This requires to manipulate TCP connection.

  ```go
  // Load CA cert
  caCertPool := x509.NewCertPool()
  if len(*caFile) != 0 {
      log.Info(*caFile)
      caCert, err := ioutil.ReadFile(*caFile)
      if err != nil {
          log.Info(trace.Wrap(err))
      } else {
          caCertPool.AppendCertsFromPEM(caCert)
      }
  }
  ```

> References

- [Adding trusted root certificates to the server.pdf](Adding trusted root certificates to the server.pdf)
- [Registries in Rancher.pdf](Registries in Rancher.pdf)
