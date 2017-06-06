# DOCKER 

- For secure connection between ETCD, we need to specify TLS to its configuration. 

  `Example (pc-core)`

  > DOCKER_OPTS="-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --tlsverify --tlscacert=/etc/docker/ca-cert.pub --tlscert=/etc/docker/pc-core.cert --tlskey=/etc/docker/pc-core.key **--cluster-advertise=enp0s8:2375 --cluster-store=etcd://pc-master:2379 --cluster-store-opt kv.cacertfile=/etc/docker/ca-cert.pub --cluster-store-opt kv.certfile=/etc/docker/pc-core.cert --cluster-store-opt kv.keyfile=/etc/docker/pc-core.key**"
  
- `ERROR: for <nodeid> Cannot start service datanode5: Error response from daemon: could not find endpoint with id 8850a2...`

  [Creating fail with Could not find container for entity id <id> after upgrading to 1.9.0 #17691](https://github.com/docker/docker/issues/17691)

  ```
  rm -f /var/lib/linkpgraph.db (or if you modified graphdb=/opt/docker : rm -rf /opt/docker/linkgraph.db)
  rm -rf /var/lib/containers/
  restart docker daemon : (for Ubuntu) sudo service docker restart
  ```
