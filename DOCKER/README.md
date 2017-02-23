# DOCKER 

For secure connection between ETCD, we need to specify TLS to its configuration. 

Example (pc-core)

> DOCKER_OPTS="-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --tlsverify --tlscacert=/etc/docker/ca-cert.pub --tlscert=/etc/docker/pc-core.cert --tlskey=/etc/docker/pc-core.key **--cluster-advertise=enp0s8:2375 --cluster-store=etcd://pc-master:2379 --cluster-store-opt kv.cacertfile=/etc/docker/ca-cert.pub --cluster-store-opt kv.certfile=/etc/docker/pc-core.cert --cluster-store-opt kv.keyfile=/etc/docker/pc-core.key**"